import { Worker } from "./workers"
import { Project } from "./projects"
import { EventEmitter } from "events"

export interface Event {
  id: string
  project: Project
  source: string
  type: string
  shortTitle?: string
  longTitle?: string
  payload?: string
  worker: Worker
}

export type EventHandler = (event: Event) => void

export class EventRegistry extends EventEmitter {

  public on(eventName: string | symbol, eventHandler: EventHandler): this {
    return super.on(eventName, eventHandler)
  }

  public fire(event: Event) {
    this.emit(`${event.source}:${event.type}`, event)
  }

}
