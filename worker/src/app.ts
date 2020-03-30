import * as events from "./brigadier/events"
import * as process from "process"
import * as brigadier from "./brigadier"
import { Logger, ContextLogger } from "./brigadier/logger"

export function run() {
  var e: events.BrignextEvent = require("/var/event/event.json")
  var w: events.BrignextWorker = require("/var/worker/worker.json")

  var exitCode: number = 0

  process.on("unhandledRejection", (reason: any, p: Promise<any>) => {
    var logger: Logger = new ContextLogger("app")
    logger.logLevel = w.logLevel
    logger.error(reason)
    exitCode = 1
  })

  process.on("exit", code => {
    if (exitCode != 0) {
      process.exit(exitCode)
    }
  })

  brigadier.fire(e, w)
}
