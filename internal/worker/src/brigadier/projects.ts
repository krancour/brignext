export interface Project {
  id: string
  kubernetes: ProjectKubernetesConfig
  secrets: { [key: string]: string }
}

export interface ProjectKubernetesConfig {
  namespace: string
}
