import { LogLevel } from "./brigadier/logger"

export interface Worker {
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
