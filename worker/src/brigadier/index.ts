import * as jobImpl from "./job"
import * as groupImpl from "./group"
import * as eventsImpl from "./events"

// events is the main event registry
export var events = new eventsImpl.EventRegistry()

export function fire(e: eventsImpl.BrignextEvent) {
  events.fire(e)
}

export class Job extends jobImpl.Job {
  public runResponse: string = "skipped run"
  public logsResponse: string = "skipped logs"

  run(): Promise<jobImpl.Result> {
      return Promise.resolve(this.runResponse)
  }

  logs(): Promise<string> {
      return Promise.resolve(this.logsResponse)
  }
}

export class Group extends groupImpl.Group {
  // This seems to be how you expose an existing class as an export.
}

