import * as jobs from "./brigadier/jobs"
import * as groups from "./brigadier/groups"
import { Event, EventRegistry } from "./brigadier/events"
import { Logger } from "./brigadier/logger"
import axios from 'axios'
import * as http2 from 'http2'

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
  logger: Logger

  constructor(
    name: string,
    image: string,
  ) {
    super(name, image)
    this.logger = new Logger(`job ${name}`)
  }

  async run(): Promise<void> {
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
    }
    catch(err) {
      // Wrap the original error to give clear context.
      throw new Error(`job ${this.name}: ${err}`)
    }
    return this.wait()
  }

  private async wait(): Promise<void> {
    return new Promise((resolve, reject) => {
      let abortMonitor = false
      let req: http2.ClientHttp2Stream
      
      let startMonitorReq = () => {
        const client = http2.connect(currentEvent.worker.apiAddress)
        client.on('error', (err: any) => console.error(err))
        req = client.request({
          ':path': `/v2/events/${currentEvent.id}/worker/jobs/${this.name}/status?watch=true`,
          "Authorization": `Bearer ${currentEvent.worker.apiToken}`
        })
        req.setEncoding('utf8')

        req.on('response', (response) => {
          let status = response[":status"]
          if (status != 200) {
            reject(new Error(`Received ${status} when attempting to stream job status`))
            abortMonitor = true
            req.destroy()
          }
        })

        req.on('data', (data: string) => {
          try {
            const status: JobStatus = JSON.parse(data)
            this.logger.log(`Job phase is ${status.phase}`)
            switch (status.phase) {
              case "ABORTED":
                reject(new Error(`Job was aborted`))
                abortMonitor = true
                req.destroy()
              case "FAILED":
                reject(new Error(`Job failed`))
                abortMonitor = true
                req.destroy()
              case "SUCCEEDED":
                resolve()
                abortMonitor = true
                req.destroy()
              case "TIMED_OUT":
                reject(new Error(`Job timed out`))
                abortMonitor = true
                req.destroy()
            }
          } catch (e) { } // Let it stay connected
        })

        req.on("end", () => {
          client.destroy()
          if (!abortMonitor) {
            // We got disconnected, but apparently not deliberately, so try
            // again.
            this.logger.log(`Had to restart the job monitor`)
            startMonitorReq()
          }
        })
      }
      startMonitorReq() // This starts the monitor for the first time.
    })
  }

  async logs(): Promise<string> {
    // // TODO: Fix this
    // if (this.cancel && this.pod == undefined || this.pod.status.phase == "Pending") {
    //   return `Job ${this.name} still unscheduled or pending when job was canceled; no logs to return.`
    // }
    // try {
    //   let result = await kubernetes.Config.defaultClient().readNamespacedPodLog(
    //     `job-${currentEvent.id}-${this.name}`,
    //     currentEvent.project.kubernetes.namespace,
    //     this.name,
    //   )
    //   return result.body
    // }
    // catch(err) {
    //   // This specifically handles errors from reading the pod logs, unpacks it,
    //   // and rethrows.
    //   throw new Error(err.response.body.message)
    // }
    return "example log"
  }

}

export class Group extends groups.Group {
  // This seems to be how you expose an existing class as an export.
}

export class Container extends jobs.Container {
  // This seems to be how you expose an existing class as an export.
}

interface JobStatus {
  phase: string
}
