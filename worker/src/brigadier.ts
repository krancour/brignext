import * as kubernetes from "@kubernetes/client-node"
import * as jobs from "./brigadier/jobs"
import * as groups from "./brigadier/groups"
import { Event, EventRegistry } from "./brigadier/events"
import { Logger } from "./brigadier/logger"
import * as request from "request"
import * as byline from "byline"
import * as k8s from "./k8s"

let currentEvent: Event

// events is the main event registry
export let events = new EventRegistry()

let fired: boolean = false
export function fire(event: Event) {
  if (!fired) {
    fired = true
    currentEvent = event
    events.fire(event)
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
      let containers = [this.primaryContainer]
      containers.push(...this.sidecarContainers)
      for (let c of containers) {
        let containerSecret = this.newContainerSecret(c)
        if (containerSecret) {
          try {
            await this.client.createNamespacedSecret(
              currentEvent.project.kubernetes.namespace,
              containerSecret,
            )
          }
          catch(err) {
            // This specifically handles errors from creating the secret,
            // unpacks it, and rethrows.
            throw new Error(err.response.body.message)
          }
        }
      }
      this.logger.log("Creating pod " + this.podName)
      let jobPod = this.newJobPod()
      try {
        await this.client.createNamespacedPod(
          currentEvent.project.kubernetes.namespace,
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

  private generateScript(container: jobs.Container): string | null {
    if (container.tasks.length == 0) {
      return null
    }
    let newCmd = "#!" + container.shell + "\n\n"

    // if shells that support the `set` command are selected, let's add some
    // sane defaults
    switch (container.shell) {
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
    if (container.tasks) {
      newCmd += container.tasks.join("\n") 
    }
    return newCmd
  }

  private newContainerSecret(container: jobs.Container): kubernetes.V1Secret {
    let script = this.generateScript(container)
    let secret = new kubernetes.V1Secret()
    secret.metadata = new kubernetes.V1ObjectMeta()
    // TODO: This naming scheme can have conflicts, so it needs to be fixed. For
    // instance, job "foo" + container "bar-bat" would have the same name as
    // job "foo-bar" + container "bat".
    secret.metadata.name = `container-${currentEvent.id}-${this.name}-${container.name}`
    secret.metadata.labels = {
      "brignext.io/component": "container",
      "brignext.io/project": currentEvent.project.id,
      "brignext.io/event": currentEvent.id,
      "brignext.io/job": this.name,
      "brignext.io/container": container.name
    }
    secret.type = "brignext.io/container"
    secret.stringData = {}
    if (script) {
      secret.stringData["main.sh"] = script
    }
    for (let key in container.env) {
      secret.stringData[key] = container.env[key]
    }
    return secret
  }

  private newJobPod(): kubernetes.V1Pod {
    let pod = new kubernetes.V1Pod()
    pod.metadata = new kubernetes.V1ObjectMeta()
    pod.metadata.name = this.podName
    pod.metadata.namespace = currentEvent.project.kubernetes.namespace
    pod.metadata.labels = {
      "brignext.io/component": "job",
      "brignext.io/project": currentEvent.project.id,
      "brignext.io/event": currentEvent.id,
      "brignext.io/job": this.name
    }

    // This is all the containers, primary first, followed by sidecars...
    let containers = [this.primaryContainer]
    containers.push(...this.sidecarContainers)

    pod.spec = new kubernetes.V1PodSpec()
    pod.spec.containers = []
    pod.spec.volumes = []

    let useSource = false
    let useDockerSocket = false
    if (currentEvent.worker.git.cloneURL) {
      // If ANY container uses source...
      for (let c of containers) {
        if (c.useSource) {
          useSource = true
        }
        if (c.docker.enabled) {
          useDockerSocket = true
        }
      }
    }

    // If ANY container needs access to source code, define a volume for the
    // source code and an init container to populate it.
    if (useSource) {
      // The volume for source code:
      pod.spec.volumes.push(
        { name: "vcs", emptyDir: {} }
      )

      // A VCS init container to populate the volume:
      let container = new kubernetes.V1Container()
      container.name = "vcs"
      container.image = "brigadecore/git-sidecar:v1.4.0"
      container.imagePullPolicy = "IfNotPresent"
      container.env = [
        { name: "BRIGADE_REMOTE_URL", value: currentEvent.worker.git.cloneURL },
        { name: "BRIGADE_COMMIT_ID", value: currentEvent.worker.git.commit },
        { name: "BRIGADE_COMMIT_REF", value: currentEvent.worker.git.ref },
        {
          name: "BRIGADE_REPO_KEY",
          valueFrom: {
            secretKeyRef: {
              name: "event-" + currentEvent.id,
              key: "gitSSHKey",
            }
          }
        },
        {
          name: "BRIGADE_REPO_SSH_CERT",
          valueFrom: {
            secretKeyRef: {
              name: "event-" + currentEvent.id,
              key: "gitSSHCert",
            }
          }
        },
        {
          name: "BRIGADE_SUBMODULES",
          value: currentEvent.worker.git.initSubmodules.toString()
        },
        { name: "BRIGADE_WORKSPACE", value: "/var/vcs" }
      ]
      let volumeMount = new kubernetes.V1VolumeMount()
      volumeMount.name = "vcs"
      volumeMount.mountPath = "/var/vcs"
      container.volumeMounts = [volumeMount]

      pod.spec.initContainers = [container]
    }

    // If ANY container wants access to the host's Docker socket AND project
    // configuration permits that, prepare a volume.
    if (useDockerSocket && currentEvent.worker.jobPolicies.allowDockerSocketMount) {
      pod.spec.volumes.push({
        name: "docker-socket",
        hostPath: {
          path: "/var/run/docker.sock"
        }
      })      
    }

    // Create the primary container AND any sidecars...
    for (let c of containers) {
      let container = new kubernetes.V1Container()
      container.volumeMounts = []
      container.name = c.name
      container.image = c.image
      container.imagePullPolicy = c.imageForcePull ? "Always" : "IfNotPresent"

      // If necessary, mount the main.sh script from a volume and execute
      // that as the container's command...
      if (c.tasks.length > 0) {
        pod.spec.volumes.push({
          name: c.name,
          secret: {
            secretName: `container-${currentEvent.id}-${this.name}-${container.name}`
          }
        })
        let jobShell = this.primaryContainer.shell ? this.primaryContainer.shell : "/bin/sh"
        container.command = [jobShell, "/var/container/main.sh"]
        container.volumeMounts.push({
          name: c.name,
          mountPath: "/var/container"
        })
      }
      if (c.args.length > 0) {
        container.args = c.args
      }

      // Add environment variables. These will ALL be secrets because some of
      // them might be secrets sourced from the project's own secrets. We don't
      // know.
      container.env = []
      for (let key in c.env) {
        container.env.push(
          {
            name: key,
            valueFrom: {
              secretKeyRef: {
                name: `container-${currentEvent.id}-${this.name}-${container.name}`,
                key: key
              }
            }
          }
        )
      }

      // If the container requests access to the host's Docker daemon AND it's
      // allowed, mount it...
      if (c.docker.enabled && currentEvent.worker.jobPolicies.allowDockerSocketMount) {
        container.volumeMounts.push({
          name: "docker-socket",
          mountPath: "/var/run/docker.sock"
        })
      }

      // If the job requests a privileged security context and it's allowed,
      // enable it...
      if (c.privileged && currentEvent.worker.jobPolicies.allowPrivileged) {
        container.securityContext = {
          privileged: true
        }
      }

      // Finally add the container to the pod spec
      pod.spec.containers.push(container)
    }

    // Jobs run once and succeed or fail. They don't restart.
    pod.spec.restartPolicy = "Never"

    // Security related settings

    // Every BrigNext project has, in its dedicated namespace, a service account
    // named "jobs", which exists for the express use of all jobs in the
    // project.
    pod.spec.serviceAccountName = "jobs"

    if (currentEvent.worker.jobPolicies.kubernetes.imagePullSecrets) { 
      pod.spec.imagePullSecrets = []
      for (let imagePullSecret of currentEvent.worker.jobPolicies.kubernetes.imagePullSecrets) {
        pod.spec.imagePullSecrets.push(
          { name: imagePullSecret }
        )
      }
    }

    // Misc. node selection settings

    // If host os is set, specify it.
    // TODO: Most BrigNext components are Linux-based, which means if a cluster
    // CAN accommodate scheduling a job to a Windows node, it MUST be a mixed
    // node cluster. In that case, there is almost certainly a taint on the
    // Windows nodes that is repelling most workloads. Therefore, there is
    // probably a toleration that needs to also be added to a Windows-based
    // job pod. Unfortunately, I don't think there's a universal approach for
    // tainting Windows nodes, so we either need to work out some specific,
    // reasonable, and well documented strategy OR we need to make the
    // toleration configurable.
    if (this.host.os) {
      pod.spec.nodeSelector = {
        "beta.kubernetes.io/os": this.host.os
      }
    }
    
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

        for (let containerStatus of this.pod.status.containerStatuses) {
          // Trap image pull errors for any container and count it as fatal
          if (
            containerStatus.state.waiting && 
            containerStatus.state.waiting.reason == "ErrImagePull"
           ) {
            this.client.deleteNamespacedPod(
              name,
              ns
            ).catch(e => this.logger.error(e.body.message))
            clearTimers()
            reject(new Error(containerStatus.state.waiting.message))
          }
          // If we're looking at the state of the primary container, try to
          // determine if we're running, succeeded, or failed...
          if (containerStatus.name == this.pod.spec.containers[0].name) {
            if (containerStatus.state.terminated) {
              if (containerStatus.state.terminated.reason == "Completed") {
                clearTimers()
                resolve()
              } else {
                clearTimers()
                reject(new Error(`Pod ${name} primary container failed to run to completion`))
              }
            }
          }
        }
        this.logger.log(
          `${currentEvent.project.kubernetes.namespace}/${this.podName} phase ${this.pod.status.phase}`
        )
        // In all other cases we fall through and let the fn be run again.
      }
      let interval = setInterval(() => {
        if (this.cancel) {
          podUpdater.abort()
          clearInterval(interval)
          clearTimeout(waiter)
          return
        }
        pollOnce(name, currentEvent.project.kubernetes.namespace, interval)
      }, 2000)
      let clearTimers = () => {
        podUpdater.abort()
        clearInterval(interval)
        clearTimeout(waiter)
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
    const url = `${k8s.config.getCurrentCluster().server}/api/v1/namespaces/${currentEvent.project.kubernetes.namespace}/pods`
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
        currentEvent.project.kubernetes.namespace,
        this.primaryContainer.name,
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

export class Container extends jobs.Container {
  // This seems to be how you expose an existing class as an export.
}
