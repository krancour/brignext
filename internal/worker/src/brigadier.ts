import * as kubernetes from "@kubernetes/client-node"
import * as jobs from "./brigadier/jobs"
import * as groups from "./brigadier/groups"
import { Event, EventRegistry } from "./brigadier/events"
import { Worker } from "./brigadier/workers"
import { Logger } from "./brigadier/logger"
import * as request from "request"
import * as byline from "byline"
import * as k8s from "./k8s"

let currentEvent: Event
let currentWorker: Worker

// events is the main event registry
export let events = new EventRegistry()

let fired: boolean = false
export function fire(event: Event, worker: Worker) {
  if (!fired) {
    fired = true
    currentEvent = event
    currentWorker = worker
    events.fire(event, worker)
  }
}

export class Job extends jobs.Job {

  podName: string
  client: kubernetes.CoreV1Api
  logger: Logger

  cancel: boolean = false
  reconnect: boolean = false

  pod: kubernetes.V1Pod

  constructor(
    name: string,
    image?: string,
    tasks?: string[],
    imageForcePull: boolean = false
  ) {
    super(name, image, tasks, imageForcePull)
    this.podName = `job-${currentEvent.id}-${name.toLowerCase()}`
    this.client = k8s.defaultClient
    this.logger = new Logger(`job ${name}`)
  }

  async run(): Promise<jobs.Result> {
    try {
      let jobSecret = this.newJobSecret()
      if (jobSecret) {
        try {
          await this.client.createNamespacedSecret(
            currentEvent.kubernetes.namespace,
            jobSecret,
          )
        }
        catch(err) {
          // This specifically handles errors from creating the secret,
          // unpacks it, and rethrows.
          throw new Error(err.response.body.message)
        }
      }
      this.logger.log("Creating pod " + this.podName)
      let jobPod = this.newJobPod(jobSecret)
      try {
        await this.client.createNamespacedPod(
          currentEvent.kubernetes.namespace,
          jobPod
        )
      }
      catch(err) {
        // This specifically handles errors from creating the pod,
        // unpacks it, and rethrows.
        throw new Error(err.response.body.message) 
      }
      await this.wait()
      let logs = await this.logs()
      return new jobs.Result(logs)
    }
    catch(err) {
      // Wrap the original error to give clear context.
      throw new Error(`job ${this.name}: ${err}`)
    }
  }

  private generateScript(): string | null {
    if (this.tasks.length == 0) {
      return null
    }
    let newCmd = "#!" + this.shell + "\n\n"

    // if shells that support the `set` command are selected, let's add some
    // sane defaults
    switch (this.shell) {
      case "/bin/sh":
        // The -e option will cause a bash script to exit immediately when a
        // command fails
        newCmd += "set -e\n\n"
        break 
      case "/bin/bash":
        // The -e option will cause a bash script to exit immediately when a
        // command fails.
        // The -o pipefail option sets the exit code of a pipeline to that of
        // the rightmost command to exit with a non-zero status, or to zero if
        // all commands of the pipeline exit successfully.
        newCmd += "set -eo pipefail\n\n"
        break
      default:
        // No-op currently
    }

    // Join the tasks to make a new command:
    if (this.tasks) {
      newCmd += this.tasks.join("\n") 
    }
    return newCmd
  }

  private newJobSecret(): kubernetes.V1Secret {
    let script = this.generateScript()
    let secret = new kubernetes.V1Secret()
    secret.metadata = new kubernetes.V1ObjectMeta()
    secret.metadata.name = this.podName
    secret.metadata.labels = {
      "brignext.io/component": "job",
      "brignext.io/project": currentEvent.projectID,
      "brignext.io/event": currentEvent.id,
      "brignext.io/job": this.name
    }
    secret.type = "brignext.io/job"
    secret.stringData = {}
    if (script) {
      secret.stringData["main.sh"] = script
    }
    for (let key in this.env) {
      secret.stringData[key] = this.env[key]
    }
    return secret
  }

  private newJobPod(
    jobSecret: kubernetes.V1Secret
  ): kubernetes.V1Pod {
    let pod = new kubernetes.V1Pod()
    pod.metadata = new kubernetes.V1ObjectMeta()
    pod.metadata.name = this.podName
    pod.metadata.namespace = currentEvent.kubernetes.namespace
    pod.metadata.labels = {
      "brignext.io/component": "job",
      "brignext.io/project": currentEvent.projectID,
      "brignext.io/event": currentEvent.id,
      "brignext.io/job": this.name
    }

    pod.spec = new kubernetes.V1PodSpec()
    pod.spec.volumes = []

    let container = new kubernetes.V1Container()
    container.volumeMounts = []

    // Conditionally describe the vcs init container
    if (this.useSource && currentWorker.git.cloneURL != "") {
      let vcsInitContainer = new kubernetes.V1Container()
      vcsInitContainer.name = "vcs"
      vcsInitContainer.image = "brigadecore/git-sidecar:v1.4.0"
      vcsInitContainer.imagePullPolicy = "Always"
      vcsInitContainer.env = [
        { name: "BRIGADE_REMOTE_URL", value: currentWorker.git.cloneURL },
        { name: "BRIGADE_COMMIT_ID", value: currentWorker.git.commit },
        { name: "BRIGADE_COMMIT_REF", value: currentWorker.git.ref },
        {
          name: "BRIGADE_REPO_KEY",
          valueFrom: {
            secretKeyRef: {
              name: "worker-" + currentEvent.id,
              key: "gitSSHKey",
            }
          }
        },
        {
          name: "BRIGADE_REPO_SSH_CERT",
          valueFrom: {
            secretKeyRef: {
              name: "worker-" + currentEvent.id,
              key: "gitSSHCert",
            }
          }
        },
        {
          name: "BRIGADE_SUBMODULES",
          value: currentWorker.git.initSubmodules.toString()
        },
        { name: "BRIGADE_WORKSPACE", value: "/var/vcs" }
      ]

      let vcsVolumeMount = new kubernetes.V1VolumeMount()
      vcsVolumeMount.name = "vcs"
      vcsVolumeMount.mountPath = "/var/vcs"

      // Init container volume mount
      vcsInitContainer.volumeMounts = [vcsVolumeMount]

      pod.spec.initContainers = [vcsInitContainer]

      // The main job container needs a similar volume mount
      container.volumeMounts.push(vcsVolumeMount)

      // Also add the volume shared by both containers to the pod spec
      pod.spec.volumes.push(
        { name: "vcs", emptyDir: {} }
      )
    }

    // Describe the main job container
    container.name = this.name
    container.image = this.image
    container.imagePullPolicy = this.imageForcePull ? "Always" : "IfNotPresent"
    if (jobSecret.stringData["main.sh"]) {
      let jobShell = this.shell
      if (jobShell == "") {
        jobShell = "/bin/sh"
      }
      container.command = [jobShell, "/var/job/main.sh"]

      container.volumeMounts.push({
        name: "job",
        mountPath: "/var/job"
      })

      pod.spec.volumes.push({
        name: "job",
        secret: {
          secretName: jobSecret.metadata.name
        }
      })
    }
    if (this.args.length > 0) {
      container.args = this.args
    }
    if (jobSecret) {
      container.env = []
      // Add environment variables
      for (let key in jobSecret.stringData) {
        container.env.push(
          {
            name: key,
            valueFrom: {
              secretKeyRef: {
                name: jobSecret.metadata.name,
                key: key
              }
            }
          }
        )
      }
    }

    // If the job requests access to the host's Docker daemon AND it's
    // allowed...
    if (this.docker.enabled && currentWorker.jobs.allowDockerSocketMount) {
      const dockerSocketVolumeName = "docker-socket"
      const dockerSocketPath = "/var/run/docker.sock"

      let dockerSocketVolumeSource = new kubernetes.V1HostPathVolumeSource()
      dockerSocketVolumeSource.path = dockerSocketPath
      let dockerSocketVolume = new kubernetes.V1Volume()
      dockerSocketVolume.name = dockerSocketVolumeName
      dockerSocketVolume.hostPath = dockerSocketVolumeSource
      pod.spec.volumes.push(dockerSocketVolume)

      let dockerSocketVolumeMount = new kubernetes.V1VolumeMount()
      dockerSocketVolumeMount.name = dockerSocketVolumeName
      dockerSocketVolumeMount.mountPath = dockerSocketPath
      container.volumeMounts.push(dockerSocketVolumeMount)
    }

    // If the job requests a privileged security context and it's allowed,
    // enable it...
    if (this.privileged && currentWorker.jobs.allowPrivileged) {
      container.securityContext = new kubernetes.V1SecurityContext()
      container.securityContext.privileged = true
    }

    // Finally add the main container to the pod spec
    pod.spec.containers = [container]

    // Jobs run once and succeed or fail. They don't restart.
    pod.spec.restartPolicy = "Never"

    // Security related settings

    // Every BrigNext project has, in its dedicated namespace, a service account
    // named "jobs", which exists for the express use of all jobs in the
    // project.
    pod.spec.serviceAccountName = "jobs"

    if (currentWorker.jobs.kubernetes.imagePullSecrets) { 
      pod.spec.imagePullSecrets = []
      for (let imagePullSecret of currentWorker.jobs.kubernetes.imagePullSecrets) {
        pod.spec.imagePullSecrets.push(
          { name: imagePullSecret }
        )
      }
    }

    // Misc. node selection settings

    // If host os is set, specify it.
    if (this.host.os) {
      pod.spec.nodeSelector = {
        "beta.kubernetes.io/os": this.host.os
      }
    }
    // TODO: This looks like something we probably shouldn't expose
    if (this.host.name) {
      pod.spec.nodeName = this.host.name
    }
    // TODO: Also something that we perhaps ought not expose
    if (this.host.nodeSelector && this.host.nodeSelector.size > 0) {
      if (!pod.spec.nodeSelector) {
        pod.spec.nodeSelector = {}
      }
      for (let k of this.host.nodeSelector.keys()) {
        pod.spec.nodeSelector[k] = this.host.nodeSelector.get(k)
      }
    }

    return pod
  }

  private async wait(): Promise<void> {
    let timeout = this.timeout || 60000
    let name = this.name
    let podUpdater: request.Request = undefined

    // This is a handle to clear the setTimeout when the promise is fulfilled.
    let waiter
    // Handle to abort the request on completion and only to ensure that we hook
    // the 'follow logs' events only once
    let followLogsRequest: request.Request = null

    this.logger.log(`Timeout set at ${timeout} milliseconds`)

    // At intervals, poll the Kubernetes server and get the pod phase. If the
    // phase is Succeeded or Failed, bail out. Otherwise, keep polling.
    //
    // The timeout sets an upper limit, and if that limit is reached, the
    // polling will be stopped.
    //
    // Essentially, we track two Timer objects: the setTimeout and the
    // setInterval. That means we have to kill both before exit, otherwise the
    // node.js process will remain running until all timeouts have executed.

    // Poll the server waiting for a Succeeded.
    let poll : Promise<void> = new Promise((resolve, reject) => {
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
          resolve()
        }
        // make sure Pod is running before we start following its logs
        else if (phase == "Running") {
          // do that only if we haven't hooked up the follow request before
          if (followLogsRequest == null && this.streamLogs) {
            followLogsRequest = followLogs(
              currentEvent.kubernetes.namespace,
              this.podName
            )
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
            this.client.deleteNamespacedPod(
              name,
              ns
            ).catch(e => this.logger.error(e.body.message))
            clearTimers()
            reject(new Error(cs[0].state.waiting.message))
          }
        }
        if (!this.streamLogs || (this.streamLogs && this.pod.status.phase != "Running")) {
          // don't display "Running" when we're asked to display job Pod logs
          this.logger.log(
            `${currentEvent.kubernetes.namespace}/${this.podName} phase ${this.pod.status.phase}`
          )
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
        pollOnce(name, currentEvent.kubernetes.namespace, interval)
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
        const url = `${k8s.config.getCurrentCluster().server}/api/v1/namespaces/${namespace}/pods/${podName}/log`
        // https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#pod-v1-core
        const requestOptions = {
          qs: {
            follow: true,
            timeoutSeconds: 200,
          },
          method: "GET",
          uri: url,
          useQuerystring: true
        }
        k8s.config.applyToRequest(requestOptions)
        const stream = new byline.LineStream()
        stream.on("data", data => {
          let logs = null
          try {
            if (data instanceof Buffer) {
              logs = data.toString()
            } else {
              logs = data
            }
            this.logger.log(
              `${currentEvent.kubernetes.namespace}/${this.podName} logs ${logs}`
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

    // This will fail if the time limit is reached.
    let timer : Promise<void> = new Promise((resolve, reject) => {
      waiter = setTimeout(() => {
        this.cancel = true
        reject(new Error("time limit (" + timeout + " ms) exceeded"))
      }, timeout)
    })

    return Promise.race([poll, timer])
  }

  private startUpdatingPod(): request.Request {
    const url = `${k8s.config.getCurrentCluster().server}/api/v1/namespaces/${currentEvent.kubernetes.namespace}/pods`
    const requestOptions = {
      qs: {
        watch: true,
        timeoutSeconds: 200,
        fieldSelector: `metadata.name=${this.podName}`
      },
      method: "GET",
      uri: url,
      useQuerystring: true,
      json: true
    }
    k8s.config.applyToRequest(requestOptions)
    const stream = new byline.LineStream()
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

  async logs(): Promise<string> {
    if (this.cancel && this.pod == undefined || this.pod.status.phase == "Pending") {
      return "pod " + this.podName + " still unscheduled or pending when job was canceled; no logs to return."
    }
    try {
      let result = await this.client.readNamespacedPodLog(
        this.podName,
        currentEvent.kubernetes.namespace
      )
      return result.body
    }
    catch(err) {
      // This specifically handles errors from reading the pod logs, unpacks it,
      // and rethrows.
      throw new Error(err.response.body.message)
    }
  }
}

export class Group extends groups.Group {
  // This seems to be how you expose an existing class as an export.
}