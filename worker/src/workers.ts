export interface Worker {
  name: string
  git: GitConfig
  jobs: JobsConfig
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