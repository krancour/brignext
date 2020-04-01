import { Job, Result } from "./jobs"

export class Group {

  public static runAll(jobs: Job[]): Promise<Result[]> {
    let g = new Group(jobs)
    return g.runAll()
  }

  public static runEach(jobs: Job[]): Promise<Result[]> {
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

  public runEach(): Promise<Result[]> {
    // TODO: Rewrite this using async/await, which will make it much cleaner.
    return this.jobs.reduce(
      (promise: Promise<Result[]>, job: Job) => {
        return promise.then((results: Result[]) => {
          return job.run().then(jobResult => {
            results.push(jobResult)
            return results
          })
        })
      },
      Promise.resolve([])
    )
  }

  public runAll(): Promise<Result[]> {
    let plist: Promise<Result>[] = []
    for (let j of this.jobs) {
      plist.push(j.run())
    }
    return Promise.all(plist)
  }

}
