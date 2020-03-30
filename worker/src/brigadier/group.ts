import * as jobImpl from "./job"

export class Group {

  public static runAll(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]> {
    let g = new Group(jobs)
    return g.runAll()
  }

  public static runEach(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]> {
    let g = new Group(jobs)
    return g.runEach()
  }

  protected jobs: jobImpl.Job[] = []
  public constructor(jobs?: jobImpl.Job[]) {
    this.jobs = jobs || []
  }

  public add(...j: jobImpl.Job[]): void {
    for (let jj of j) {
      this.jobs.push(jj)
    }
  }

  public length(): number {
    return this.jobs.length
  }

  public runEach(): Promise<jobImpl.Result[]> {
    // TODO: Rewrite this using async/await, which will make it much cleaner.
    return this.jobs.reduce(
      (promise: Promise<jobImpl.Result[]>, job: jobImpl.Job) => {
        return promise.then((results: jobImpl.Result[]) => {
          return job.run().then(jobResult => {
            results.push(jobResult)
            return results
          })
        })
      },
      Promise.resolve([])
    )
  }

  public runAll(): Promise<jobImpl.Result[]> {
    let plist: Promise<jobImpl.Result>[] = []
    for (let j of this.jobs) {
      plist.push(j.run())
    }
    return Promise.all(plist)
  }

}
