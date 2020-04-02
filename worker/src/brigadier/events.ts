import { EventEmitter } from "events"

export interface Event {
  id: string
  projectID: string
  provider: string
  type: string
  shortTitle?: string
  longTitle?: string
  kubernetes: EventKubernetesConfig
}

export interface EventKubernetesConfig {
  namespace: string
}

export type EventHandler = (event: Event) => void

export class EventRegistry extends EventEmitter {
  
  public on(eventName: string | symbol, eventHandler: EventHandler): this {
    return super.on(eventName, eventHandler)
  }

  public fire(event: Event) {
    this.emit(`${event.provider}:${event.type}`, event)
  }

}