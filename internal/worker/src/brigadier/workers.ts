export interface Worker {
  git: GitConfig
  jobs: JobsSpec
  configFilesDirectory: string
}

export interface GitConfig {
  cloneURL: string
  commit: string
  ref: string
  initSubmodules: boolean
}

export interface JobsSpec {
  allowPrivileged: boolean
  allowDockerSocketMount: boolean
  kubernetes: JobsKubernetesConfig
}

export interface JobsKubernetesConfig {
  imagePullSecrets: string[]
}
