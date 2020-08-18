export interface Project {
  id: string
  // TODO: The kubernetes field can go away when we transition to using the API
  // to watch job status
  kubernetes: ProjectKubernetesConfig
  secrets: Map<string, string>
}

export interface ProjectKubernetesConfig {
  namespace: string
}
