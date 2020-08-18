import * as kubernetes from "@kubernetes/client-node"
import * as jobs from "./brigadier/jobs"
import * as groups from "./brigadier/groups"
import { Event, EventRegistry } from "./brigadier/events"
import { Logger } from "./brigadier/logger"
import * as request from "request"
import * as byline from "byline"
import * as k8s from "./k8s"
import axios from 'axios';

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
  k8sClient: kubernetes.CoreV1Api
  logger: Logger

  cancel: boolean = false
  reconnect: boolean = false

  pod: kubernetes.V1Pod

  constructor(
    name: string,
    image: string,
  ) {
    super(name, image)
    this.podName = `job-${currentEvent.id}-${name}`
    this.k8sClient = k8s.defaultClient
    this.logger = new Logger(`job ${name}`)
  }

  async run(): Promise<jobs.Result> {
    this.logger.log(`Creating job ${this.name}`)
    try {
      let response = await axios({
        method: "put",
        url: `${currentEvent.worker.apiAddress}/v2/events/${currentEvent.id}/worker/jobs/${this.name}/spec`,
        headers: {
          Authorization: `Bearer ${currentEvent.worker.apiToken}`
        },
        data: {
          apiVersion: "github.com/krancour/brignext/v2",
          kind: "JobSpec",
          primaryContainer: this.primaryContainer,
          sidecarContainers: this.sidecarContainers,
          timeoutSeconds: this.timeout,
          host: this.host
        },
      })
      if (response.status != 201) {
        throw new Error(response.data)
      }
      // TODO: This watches directly using k8s-- don't do this
      await this.wait()
      let logs = await this.logs()
      return new jobs.Result(logs)
    }
    catch(err) {
      // Wrap the original error to give clear context.
      throw new Error(`job ${this.name}: ${err}`)
    }
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
            this.k8sClient.deleteNamespacedPod(
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
      let result = await this.k8sClient.readNamespacedPodLog(
        this.podName,
        currentEvent.project.kubernetes.namespace,
        this.name,
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
