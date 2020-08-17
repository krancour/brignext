export interface Worker {
  apiToken: string
  configFilesDirectory: string
  defaultConfigFiles: { [key: string]: string }
}
