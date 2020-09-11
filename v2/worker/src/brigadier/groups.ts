import { Job } from "./jobs"

export class Group {

  public static async runAll(jobs: Job[]): Promise<void[]> {
    let g = new Group(jobs)
    return g.runAll()
  }

  public static async runEach(jobs: Job[]): Promise<void> {
    let g = new Group(jobs)
    return g.runEach()
  }

  protected jobs: Job[] = []

  public constructor(jobs?: Job[]) {
    this.jobs = jobs || []
  }

  public add(...j: Job[]): void {
    for (let jj of j) {
      this.jobs.push(jj)
    }
  }

  public length(): number {
    return this.jobs.length
  }

  public async runEach(): Promise<void> {
    for (let job of this.jobs) {
      await job.run()
    }
  }

  public async runAll(): Promise<void[]> {
    let plist: Promise<void>[] = []
    for (let j of this.jobs) {
      plist.push(j.run())
    }
    return Promise.all(plist)
  }

}
