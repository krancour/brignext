import * as jobs from "./jobs"
import * as groups from "./groups"
import { EventRegistry } from "./events"

// events is the main event registry
export let events = new EventRegistry()

export class Job extends jobs.Job {

  run(): Promise<jobs.Result> {
    return Promise.resolve(new jobs.Result("skipped run"))
  }

  logs(): Promise<string> {
    return Promise.resolve("skipped logs")
  }
  
}

export class Group extends groups.Group {
  // This seems to be how you expose an existing class as an export.
}
