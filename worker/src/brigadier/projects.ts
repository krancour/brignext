export interface Project {
  id: string
  kubernetes: ProjectKubernetesConfig
  secrets: Map<string, string>
}

export interface ProjectKubernetesConfig {
  namespace: string
}
