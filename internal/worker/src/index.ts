import * as fs from "fs"
import * as moduleAlias from "module-alias"
import * as path from "path"
import * as requireFromString from "require-from-string"

import { Logger } from "./brigadier/logger"
import { Event } from "./brigadier/events"
import * as brigadier from "./brigadier"

const logger = new Logger([])
const version = require("../package.json").version
logger.log(`brignext-worker version: ${version}`)

const event: Event = require("/var/event/event.json")

let script = ""
let scriptPath = path.join("/var/vcs", event.worker.configFilesDirectory, "brignext.js")
if (fs.existsSync(scriptPath)) {
  script = fs.readFileSync(scriptPath, "utf8")
} else {
  script = event.worker.defaultConfigFiles["brignext.js"]
}

if (script) {
  // Install aliases for common ways of referring to BrigNext/Brigadier.
  moduleAlias.addAliases({
    "brignext": __dirname + "/brigadier",
    "brigadier": __dirname + "/brigadier",
    "@brigadecore/brigadier": __dirname + "/brigadier",
  })

  // Add the current module resolution paths to module-alias, so the
  // node_modules that prestart.js adds to will be resolvable from the BrigNext
  // script and any local dependencies.
  module.paths.forEach(moduleAlias.addPath)

  moduleAlias()
  requireFromString(script)
}

let exitCode: number = 0

process.on("unhandledRejection", (reason: any, p: Promise<any>) => {
  logger.error(reason)
  exitCode = 1
})

process.on("exit", code => {
  if (exitCode != 0) {
    process.exit(exitCode)
  }
})

brigadier.fire(event)
