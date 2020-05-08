export interface Worker {
  git: GitConfig
  jobsConfig: JobsConfig
  configFilesDirectory: string
}

export interface GitConfig {
  cloneURL: string
  commit: string
  ref: string
  initSubmodules: boolean
}

export interface JobsConfig {
  allowPrivileged: boolean
  allowDockerSocketMount: boolean
  kubernetes: JobsKubernetesConfig
}

export interface JobsKubernetesConfig {
  imagePullSecrets: string[]
}
