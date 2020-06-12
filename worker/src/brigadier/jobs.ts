const defaultShell: string = "/bin/sh"

const defaultTimeout: number = 1000 * 60 * 15

const defaultJobImage: string = "debian:jessie-slim"

export class Result {
  data: string
  constructor(msg: string) {
    this.data = msg
  }
  toString(): string {
    return this.data
  }
}

export class JobHost {
  public os?: string
  public nodeSelector: Map<string, string> = new Map<string, string>()
}

export class JobDockerMount {
  public enabled: boolean = false
}

export class Container {
  public name: string
  public shell: string = defaultShell
  public tasks: string[]
  public args: string[] = []
  public env: { [key: string]: string } = {}
  public image: string = defaultJobImage
  public imageForcePull: boolean = false
  public mountPath: string = "/src"
  public timeout: number = defaultTimeout
  public useSource: boolean = true
  public privileged: boolean = false
  public docker: JobDockerMount = new JobDockerMount()

  constructor(
    name: string,
    image?: string,
    tasks?: string[],
    imageForcePull: boolean = false
  ) {
    if (!jobNameIsValid(name)) {
      throw new Error(
        "container name must be lowercase letters, numbers, and '-', and must not start or end with '-', having max length " +
        Job.MAX_JOB_NAME_LENGTH
      )
    }
    this.name = name.toLocaleLowerCase()
    this.image = image || ""
    this.tasks = tasks || []
    this.imageForcePull = imageForcePull
  }
}

export abstract class Job {
  public static readonly MAX_JOB_NAME_LENGTH = 36
  public name: string
  public primaryContainer: Container
  public sidecarContainers: Container[] = []
  public timeout: number = defaultTimeout
  public host: JobHost = new JobHost()

  constructor(
    name: string,
    image?: string,
    tasks?: string[],
    imageForcePull: boolean = false
  ) {
    if (!jobNameIsValid(name)) {
      throw new Error(
        "job name must be lowercase letters, numbers, and '-', and must not start or end with '-', having max length " +
        Job.MAX_JOB_NAME_LENGTH
      )
    }
    this.name = name.toLocaleLowerCase()
    this.primaryContainer = new Container(name, image, tasks, imageForcePull)
  }

  public abstract run(): Promise<Result>

  public abstract logs(): Promise<string>
}

function jobNameIsValid(name: string): boolean {
  return (
    name.length <= Job.MAX_JOB_NAME_LENGTH &&
    /^(([a-z0-9][-a-z0-9.]*)?[a-z0-9])+$/.test(name)
  )
}
