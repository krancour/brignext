const defaultTimeout: number = 1000 * 60 * 15

export abstract class Job {
  public name: string
  public primaryContainer: Container
  public sidecarContainers: Map<string, Container>
  public timeout: number = defaultTimeout
  public host: JobHost = new JobHost()

  constructor(
    name: string,
    image: string,
  ) {
    this.name = name
    this.primaryContainer = new Container(image)
  }

  public abstract run(): Promise<Result>

  public abstract logs(): Promise<string>
}

export class Container {
  public image: string
  public imagePullPolicy: string = "IfNotPresent"
  public command: string[] = []
  public arguments: string[] = []
  public environment: Map<string, string> = new Map<string, string>()
  public useWorkspace: boolean = false
  public workspaceMountPath: string = "/var/workspace"
  public useSource: boolean = false
  public sourceMountPath: string = "/var/vcs"
  public privileged: boolean = false
  public useHostDockerSocket: boolean = false

  constructor(image: string) {
    this.image = image
  }
}

export class JobHost {
  public os?: string
  public nodeSelector: Map<string, string> = new Map<string, string>()
}

export class Result {
  data: string
  constructor(msg: string) {
    this.data = msg
  }
  toString(): string {
    return this.data
  }
}
