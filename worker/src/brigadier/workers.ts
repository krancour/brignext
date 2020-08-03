export interface Worker {
  git: GitConfig
  jobPolicies: JobPolicies
  configFilesDirectory: string
  defaultConfigFiles: { [key: string]: string }
}

export interface GitConfig {
  cloneURL: string
  commit: string
  ref: string
  initSubmodules: boolean
}

export interface JobPolicies {
  allowPrivileged: boolean
  allowDockerSocketMount: boolean
  kubernetes: KubernetesJobPolicies
}

export interface KubernetesJobPolicies {
  imagePullSecrets: string[]
}
