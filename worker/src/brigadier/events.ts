import { EventEmitter } from "events"
import { LogLevel } from "./logger"

export interface BrignextEvent {
  id: string
  projectID: string
  provider: string
  type: string
  shortTitle?: string
  longTitle?: string
  kubernetes: EventKubernetesConfig
  cause?: Cause
}

export interface EventKubernetesConfig {
  namespace: string
}

export interface BrignextWorker {
  name: string
  git: GitConfig
  jobs: JobsConfig
  logLevel: LogLevel
}

export interface GitConfig {
  cloneURL: string
  commit: string
  ref: string
  initSubmodules: boolean
}

export interface JobsConfig {
  allowPrivileged: boolean
  allowHostMounts: boolean
  kubernetes: JobsKubernetesConfig
}

export interface JobsKubernetesConfig {
  imagePullSecrets: string
}

export interface Cause {
  event?: BrignextEvent
  reason?: any
  trigger?: string
}

export type EventHandler = (e: BrignextEvent) => void

export class EventRegistry extends EventEmitter {

  constructor() {
    super()
    this.on("ping", (e: BrignextEvent) => {
      console.log("ping")
    })
  }

  public has(name: string) {
    return this.listenerCount(name) > 0
  }

  public fire(e: BrignextEvent) {
    this.emit(e.type, e)
  }

  public on(eventName: string | symbol, cb: EventHandler): this {
    return super.on(eventName, cb)
  }

}
