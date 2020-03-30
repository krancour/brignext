const defaultShell: string = "/bin/sh"

const defaultTimeout: number = 1000 * 60 * 15

const brigadeImage: string = "debian:jessie-slim"

export const dockerSocketMountPath = "/var/run/docker.sock"
export const dockerSocketMountName = "docker-socket"

export interface JobRunner {
  start(): Promise<JobRunner>
  wait(): Promise<Result>
}

export interface Result {
  toString(): string
}

export class JobHost {
  public name?: string
  public os?: string
  public nodeSelector: Map<string, string>

  constructor() {
    this.nodeSelector = new Map<string, string>()
  }
}

export class JobDockerMount {
  public enabled: boolean = false
}

export abstract class Job {
  public static readonly MAX_JOB_NAME_LENGTH = 36
  public name: string
  public shell: string = defaultShell
  public tasks: string[]
  public args: string[]
  public env: { [key: string]: string }
  public image: string = brigadeImage
  public imageForcePull: boolean = false
  public imagePullSecrets: string[] = []
  public mountPath: string = "/src"
  public timeout: number = defaultTimeout
  public useSource: boolean = true
  public privileged: boolean = false
  public host: JobHost
  public docker: JobDockerMount
  public annotations: { [key: string]: string } = {}
  public streamLogs: boolean = false

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
    this.image = image || ""
    this.imageForcePull = imageForcePull
    this.tasks = tasks || []
    this.args = []
    this.env = {}
    this.docker = new JobDockerMount()
    this.host = new JobHost()
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
