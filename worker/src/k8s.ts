import * as kubernetes from "@kubernetes/client-node"
import * as jobs from "./brigadier/job"
import { ContextLogger } from "./brigadier/logger"
import { BrignextEvent, BrignextWorker } from "./brigadier/events"
import * as fs from "fs"
import * as path from "path"
import * as request from "request"
import * as byline_1 from "byline"

const defaultClient = kubernetes.Config.defaultClient()

const retry = (fn, args, delay, times) => {
  // exponential back-off retry if status is in the 500s
  return fn.apply(defaultClient, args).catch(err => {
    if (
      times > 0 &&
      err.response &&
      500 <= err.response.statusCode &&
      err.response.statusCode < 600
    ) {
      return new Promise(resolve => {
        setTimeout(() => {
          resolve(retry(fn, args, delay * 2, times - 1))
        }, delay)
      })
    }
    return Promise.reject(err)
  })
}

const wrapClient = fns => {
  // wrap client methods with retry logic
  for (let fn of fns) {
    let originalFn = defaultClient[fn.name]
    defaultClient[fn.name] = function () {
      return retry(originalFn, arguments, 4000, 5)
    }
  }
}

wrapClient([
  defaultClient.readNamespacedPodLog,
  defaultClient.createNamespacedSecret,
  defaultClient.createNamespacedPod,
  defaultClient.deleteNamespacedPod
]) 

const getKubeConfig = (): kubernetes.KubeConfig => {
  const kc = new kubernetes.KubeConfig()
  const config =
    process.env.KUBECONFIG || path.join(process.env.HOME, ".kube", "config")
  if (fs.existsSync(config)) {
    kc.loadFromFile(config)
  } else {
    kc.loadFromCluster()
  }
  return kc
}

const kc = getKubeConfig()

class K8sResult implements jobs.Result {
  data: string
  constructor(msg: string) {
    this.data = msg
  }
  toString(): string {
    return this.data
  }
}

// JobRunner provides a Kubernetes implementation of the JobRunner interface.
export class JobRunner implements jobs.JobRunner {
  name: string
  
  event: BrignextEvent
  worker: BrignextWorker
  job: jobs.Job
  
  client: kubernetes.CoreV1Api
  logger: ContextLogger

  runner: kubernetes.V1Pod

  cancel: boolean = false
  reconnect: boolean = false

  // TODO: This `pod` attribute seems to be for reflecting current status of
  // the worker pod. This stands in contrast to the `runner` attrbiute, whose
  // role seems only to have to do with defining / creating the worker pod.
  pod: kubernetes.V1Pod

  constructor(e: BrignextEvent, w: BrignextWorker, job: jobs.Job) {
    this.name = `${this.event.id}-${this.worker.name}`

    this.event = e
    this.worker = w
    this.job = job

    this.client = defaultClient
    this.logger = new ContextLogger("k8s", w.logLevel)

    this.runner = newRunnerPod(
      this.name,
      this.event,
      this.worker,
      this.job
    )

    // TODO: A lot of this and a lot of what follows can probably be moved into
    // the newRunnerPod function.
    let envVars: kubernetes.V1EnvVar[] = []
    for (let key in job.env) {
      let val = job.env[key]
      envVars.push(
        {
          name: key,
          value: val
        }
      )
    }

    this.runner.spec.containers[0].env = envVars

    this.runner.spec.initContainers = []
    if (job.useSource && this.worker.git.cloneURL != "") {
      // Add the VCS init container.
      this.runner.spec.initContainers = [
        vcsInitContainerSpec(e, w, "/src")
      ]

      // Add volume/volume mounts
      this.runner.spec.volumes.push(
        { name: "vcs", emptyDir: {} } as kubernetes.V1Volume
      )
      this.runner.spec.containers[0].volumeMounts.push(
        { name: "vcs", mountPath: "/var/vcs" } as kubernetes.V1VolumeMount
      )
    }

    if (job.imagePullSecrets) {
      this.runner.spec.imagePullSecrets = []
      for (let secret of job.imagePullSecrets) {
        this.runner.spec.imagePullSecrets.push({ name: secret })
      }
    }

    // If host os is set, specify it.
    if (job.host.os) {
      this.runner.spec.nodeSelector = {
        "beta.kubernetes.io/os": job.host.os
      }
    }
    if (job.host.name) {
      this.runner.spec.nodeName = job.host.name
    }
    if (job.host.nodeSelector && job.host.nodeSelector.size > 0) {
      if (!this.runner.spec.nodeSelector) {
        this.runner.spec.nodeSelector = {}
      }
      for (const k of job.host.nodeSelector.keys()) {
        this.runner.spec.nodeSelector[k] = job.host.nodeSelector.get(k)
      }
    }

    // If the job needs access to a docker daemon, mount in the host's docker socket
    if (job.docker.enabled && this.worker.jobs.allowHostMounts) {
      var dockerVol = new kubernetes.V1Volume()
      var dockerMount = new kubernetes.V1VolumeMount()
      var hostPath = new kubernetes.V1HostPathVolumeSource()
      hostPath.path = jobs.dockerSocketMountPath
      dockerVol.name = jobs.dockerSocketMountName
      dockerVol.hostPath = hostPath
      dockerMount.name = jobs.dockerSocketMountName
      dockerMount.mountPath = jobs.dockerSocketMountPath
      this.runner.spec.volumes.push(dockerVol)
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].volumeMounts.push(dockerMount)
      }
    }

    if (job.args.length > 0) {
      this.runner.spec.containers[0].args = job.args
    }

    let newCmd = generateScript(job)
    if (!newCmd) {
      this.runner.spec.containers[0].command = null
    } else {
      this.secret.data["main.sh"] = b64enc(newCmd)
    }

    // If the job asks for privileged mode and the project allows this, enable it.
    if (job.privileged && this.worker.jobs.allowPrivileged) {
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].securityContext.privileged = true
      }
    }
    return this
  }

  public logs(): Promise<string> {
    let podName = this.name
    let k = this.client
    if (this.cancel && this.pod == undefined || this.pod.status.phase == "Pending") {
      return Promise.resolve<string>(
        "pod " + podName + " still unscheduled or pending when job was canceled; no logs to return.")
    }
    return Promise.resolve<string>(
      k.readNamespacedPodLog(podName, this.event.kubernetes.namespace).then(result => {
        return result.body
      })
    )
  }

  /**
   * run starts a job and then waits until it is running.
   *
   * The Promise it returns will return when the pod is either marked
   * Success (resolve) or Failure (reject)
   */
  public run(): Promise<jobs.Result> {
    return this.start()
      .then(r => r.wait())
      .then(r => {
        return this.logs()
      })
      .then(response => {
        return new K8sResult(response)
      })
  }

  /** start begins a job, and returns once it is scheduled to run.*/
  public start(): Promise<jobs.JobRunner> {
    // Now we have pod and a secret defined. Time to create them.

    let k = this.client

    return new Promise((resolve, reject) => {
      k.createNamespacedSecret(this.event.kubernetes.namespace, this.secret)
        .then(result => {
          this.logger.log("Creating pod " + this.runner.metadata.name)
          // Once namespace creation has been accepted, we create the pod.
          return k.createNamespacedPod(this.event.kubernetes.namespace, this.runner)
        })
        .then(result => {
          resolve(this)
        })
        .catch(reason => {
          reject(new Error(reason.body.message))
        })
    })
  }

  /**
   * update pod info on event using watch
   */
  private startUpdatingPod(): request.Request {
    const url = `${kc.getCurrentCluster().server}/api/v1/namespaces/${this.event.kubernetes.namespace}/pods`
    const requestOptions = {
      qs: {
        watch: true,
        timeoutSeconds: 200,
        labelSelector: `project=${this.event.projectID},event=${this.event.id},worker=${this.worker.name},job=${this.job.name}`
      },
      method: "GET",
      uri: url,
      useQuerystring: true,
      json: true
    }
    kc.applyToRequest(requestOptions)
    const stream = new byline_1.LineStream()
    stream.on("data", data => {
      let obj = null
      try {
        if (data instanceof Buffer) {
          obj = JSON.parse(data.toString())
        } else {
          obj = JSON.parse(data)
        }
      } catch (e) { } //let it stay connected.
      if (obj && obj.object) {
        this.pod = obj.object as kubernetes.V1Pod
      }
    })
    const req = request(requestOptions, (error, response, body) => {
      if (error) {
        this.logger.error(error.body.message)
        this.reconnect = true //reconnect unless aborted
      }
    })
    req.pipe(stream)
    req.on("end", () => {
      this.reconnect = true
    }) //stay connected on transient faults
    return req
  }

  /** wait listens for the running job to complete.*/
  public wait(): Promise<jobs.Result> {
    // Should probably protect against the case where start() was not called
    let k = this.client
    let timeout = this.job.timeout || 60000
    let name = this.name
    let podUpdater: request.Request = undefined

    // This is a handle to clear the setTimeout when the promise is fulfilled.
    let waiter
    // Handle to abort the request on completion and only to ensure that we hook the 'follow logs' events only once
    let followLogsRequest: request.Request = null

    this.logger.log(`Timeout set at ${timeout} milliseconds`)

    // At intervals, poll the Kubernetes server and get the pod phase. If the
    // phase is Succeeded or Failed, bail out. Otherwise, keep polling.
    //
    // The timeout sets an upper limit, and if that limit is reached, the
    // polling will be stopped.
    //
    // Essentially, we track two Timer objects: the setTimeout and the setInterval.
    // That means we have to kill both before exit, otherwise the node.js process
    // will remain running until all timeouts have executed.

    // Poll the server waiting for a Succeeded.
    let poll = new Promise((resolve, reject) => {
      let pollOnce = (name, ns, i) => {
        if (!podUpdater) {
          podUpdater = this.startUpdatingPod()
        } else if (!this.cancel && this.reconnect) {
          //if not intentionally cancelled, reconnect
          this.reconnect = false
          try {
            podUpdater.abort()
          } catch (e) {
            this.logger.log(e)
          }
          podUpdater = this.startUpdatingPod()
        }
        if (!this.pod || this.pod.status == undefined) {
          this.logger.log("Pod not yet scheduled")
          return
        }

        let phase = this.pod.status.phase
        if (phase == "Succeeded") {
          clearTimers()
          let result = new K8sResult(phase)
          resolve(result)
        }
        // make sure Pod is running before we start following its logs
        else if (phase == "Running") {
          // do that only if we haven't hooked up the follow request before
          if (followLogsRequest == null && this.job.streamLogs) {
            followLogsRequest = followLogs(this.pod.metadata.namespace, this.pod.metadata.name)
          }
        } else if (phase == "Failed") {
          clearTimers()
          reject(new Error(`Pod ${name} failed to run to completion`))
        } else if (phase == "Pending") {
          // Trap image pull errors and consider them fatal.
          let cs = this.pod.status.containerStatuses
          if (
            cs &&
            cs.length > 0 &&
            cs[0].state.waiting &&
            cs[0].state.waiting.reason == "ErrImagePull"
          ) {
            k.deleteNamespacedPod(
              name,
              ns,
              "true",
              new kubernetes.V1DeleteOptions()
            ).catch(e => this.logger.error(e.body.message))
            clearTimers()
            reject(new Error(cs[0].state.waiting.message))
          }
        }
        if (!this.job.streamLogs || (this.job.streamLogs && this.pod.status.phase != "Running")) {
          // don't display "Running" when we're asked to display job Pod logs
          this.logger.log(`${this.pod.metadata.namespace}/${this.pod.metadata.name} phase ${this.pod.status.phase}`)
        }
        // In all other cases we fall through and let the fn be run again.
      }
      let interval = setInterval(() => {
        if (this.cancel) {
          podUpdater.abort()
          clearInterval(interval)
          clearTimeout(waiter)
          return
        }
        pollOnce(name, this.event.kubernetes.namespace, interval)
      }, 2000)
      let clearTimers = () => {
        podUpdater.abort()
        clearInterval(interval)
        clearTimeout(waiter)
        if (followLogsRequest != null) {
          followLogsRequest.abort()
        }
      }

      // follows logs for the specified namespace/Pod combination
      let followLogs = (namespace: string, podName: string): request.Request => {
        const url = `${kc.getCurrentCluster().server}/api/v1/namespaces/${namespace}/pods/${podName}/log`
        //https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#pod-v1-core
        const requestOptions = {
          qs: {
            follow: true,
            timeoutSeconds: 200,
          },
          method: "GET",
          uri: url,
          useQuerystring: true
        }
        kc.applyToRequest(requestOptions)
        const stream = new byline_1.LineStream()
        stream.on("data", data => {
          let logs = null
          try {
            if (data instanceof Buffer) {
              logs = data.toString()
            } else {
              logs = data
            }
            this.logger.log(
              `${this.pod.metadata.namespace}/${this.pod.metadata.name} logs ${logs}`
            )
          } catch (e) { } //let it stay connected.
        })
        const req = request(requestOptions, (error, response, body) => {
          if (error) {
            if (error.body) {
              this.logger.error(error.body.message)
            }
            this.reconnect = true //reconnect unless aborted
          }
        })
        req.pipe(stream)
        return req
      }

    })

    // This will fail if the timelimit is reached.
    let timer = new Promise((solve, reject) => {
      waiter = setTimeout(() => {
        this.cancel = true
        reject(new Error("time limit (" + timeout + " ms) exceeded"))
      }, timeout)
    })

    return Promise.race([poll, timer])
  }

}

function vcsInitContainerSpec(
  e: BrignextEvent,
  w: BrignextWorker,
  local: string
): kubernetes.V1Container {
  let spec = new kubernetes.V1Container()
  spec.name = "vcs"
  spec.env = [
    envVar("BRIGADE_REMOTE_URL", w.git.cloneURL),
    envVar("BRIGADE_COMMIT_ID", w.git.commit),
    envVar("BRIGADE_COMMIT_REF", w.git.ref),
    envVar("BRIGADE_WORKSPACE", "/var/vcs"),
    envVar("BRIGADE_SUBMODULES", w.git.initSubmodules.toString())
  ]
  spec.image = "brigadecore/git-sidecar:latest"
  spec.imagePullPolicy = "Always"
  spec.volumeMounts = [volumeMount("vcs", local)]
  return spec
}

function newRunnerPod(
  podname: string,
  e: BrignextEvent,
  w: BrignextWorker,
  j: jobs.Job
): kubernetes.V1Pod {
  let pod = new kubernetes.V1Pod()
  pod.metadata = new kubernetes.V1ObjectMeta()
  pod.metadata.name = podname
  pod.metadata.labels = {
    "brignext.io/component": "job",
    "brignext.io/project": e.projectID,
    "brignext.io/event": e.id,
    "brignext.io/worker": w.name,
    "brignext.io/job": "job"
  }
  pod.metadata.annotations = j.annotations

  let c1 = new kubernetes.V1Container()
  c1.name = "brigaderun"
  c1.image = j.image

  let jobShell = j.shell
  if (jobShell == "") {
    jobShell = "/bin/sh"
  }
  c1.command = [jobShell, "/hook/main.sh"]

  c1.imagePullPolicy = j.imageForcePull ? "Always" : "IfNotPresent"
  c1.securityContext = new kubernetes.V1SecurityContext()

  
  pod.spec = new kubernetes.V1PodSpec()
  pod.spec.containers = [c1]
  pod.spec.restartPolicy = "Never"
  pod.spec.serviceAccount = "jobs"
  pod.spec.serviceAccountName = "jobs"
  return pod
}

function envVar(key: string, value: string): kubernetes.V1EnvVar {
  let e = new kubernetes.V1EnvVar()
  e.name = key
  e.value = value
  return e
}

function volumeMount(
  name: string,
  mountPath: string
): kubernetes.V1VolumeMount {
  let v = new kubernetes.V1VolumeMount()
  v.name = name
  v.mountPath = mountPath
  return v
}

export function b64enc(original: string): string {
  return Buffer.from(original).toString("base64")
}

export function b64dec(encoded: string): string {
  return Buffer.from(encoded, "base64").toString("utf8")
}

function generateScript(job: jobs.Job): string | null {
  if (job.tasks.length == 0) {
    return null
  }
  let newCmd = "#!" + job.shell + "\n\n"

  // if shells that support the `set` command are selected, let's add some sane defaults
  switch (job.shell) {
    case "/bin/sh":
      // The -e option will cause a bash script to exit immediately when a command fails
      newCmd += "set -e\n\n"
      break 
    case "/bin/bash":
      // The -e option will cause a bash script to exit immediately when a command fails
      // The -o pipefail option sets the exit code of a pipeline to that of the rightmost command
      // to exit with a non-zero status, or to zero if all commands of the pipeline exit successfully.
      newCmd += "set -eo pipefail\n\n"
      break
    default:
      // No-op currently
  }

  // Join the tasks to make a new command:
  if (job.tasks) {
    newCmd += job.tasks.join("\n") 
  }
  return newCmd
}
