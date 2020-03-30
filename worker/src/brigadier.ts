import * as jobImpl from "./brigadier/job"
import * as groupImpl from "./brigadier/group"
import * as eventsImpl from "./brigadier/events"
import { JobRunner } from "./k8s"

var currentEvent = null
var currentWorker = null

// events is the main event registry
export var events = new eventsImpl.EventRegistry()

export function fire(e: eventsImpl.BrignextEvent, w: eventsImpl.BrignextWorker) {
  currentEvent = e
  currentWorker = w
  events.fire(e)
}

export class Job extends jobImpl.Job {
  jr: JobRunner

  run(): Promise<jobImpl.Result> {
    this.jr = new JobRunner(currentEvent, currentWorker, this)
    return this.jr.run().catch(err => {
      // Wrap the message to give clear context.
      console.error(err)
      let msg = `job ${ this.name }(${this.jr.name}): ${err}`
      return Promise.reject(new Error(msg))
    })
  }

  logs(): Promise<string> {
    return this.jr.logs()
  }
}

export class Group extends groupImpl.Group {
  // This seems to be how you expose an existing class as an export.
}
