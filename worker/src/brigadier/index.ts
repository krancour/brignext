import * as jobs from "./jobs"
import * as groups from "./groups"
import { EventRegistry } from "./events"

// events is the main event registry
export let events = new EventRegistry()

export class Job extends jobs.Job {

  run(): Promise<void> {
    return Promise.resolve()
  }

  logs(): Promise<string> {
    return Promise.resolve("skipped logs")
  }

}

export class Group extends groups.Group {
  // This seems to be how you expose an existing class as an export.
}
