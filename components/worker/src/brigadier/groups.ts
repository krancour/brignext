import { Job, Result } from "./jobs"

export class Group {

  public static async runAll(jobs: Job[]): Promise<Result[]> {
    let g = new Group(jobs)
    return g.runAll()
  }

  public static async runEach(jobs: Job[]): Promise<Result[]> {
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

  public async runEach(): Promise<Result[]> {
    let results: Result[] = []
    for (let job of this.jobs) {
      let result = await job.run()
      results.push(result)
    }
    return results
  }

  public async runAll(): Promise<Result[]> {
    let plist: Promise<Result>[] = []
    for (let j of this.jobs) {
      plist.push(j.run())
    }
    return Promise.all(plist)
  }

}
