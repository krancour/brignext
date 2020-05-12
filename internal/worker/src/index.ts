import * as fs from "fs"
import * as moduleAlias from "module-alias"
import * as path from "path"

import { Logger } from "./brigadier/logger"
import { Event } from "./brigadier/events"
import { Worker } from "./brigadier/workers"
import * as brigadier from "./brigadier"

const logger = new Logger([])
const version = require("../package.json").version
logger.log(`brignext-worker version: ${version}`)

const event: Event = require("/var/event/event.json")
const worker: Worker = require("/var/worker/worker.json")

const scriptLocations = [
  "/var/vcs/" + worker.configFilesDirectory + "/brigade.js",
  "/var/worker/brigade.js"
]

let script = ""
for (let scriptFileLocation of scriptLocations) {
  if (fs.existsSync(scriptFileLocation)) {
    script = scriptFileLocation
  }
}

if (script) {
  // Install aliases for common ways of referring to Brigade/Brigadier.
  moduleAlias.addAliases({
    "brigade": __dirname + "/brigadier",
    "brigadier": __dirname + "/brigadier",
    "@brigadecore/brigadier": __dirname + "/brigadier",
  })

  // Add the current module resolution paths to module-alias, so the
  // node_modules that prestart.js adds to will be resolvable from the Brigade
  // script and any local dependencies.
  module.paths.forEach(moduleAlias.addPath)

  const realScriptPath = fs.realpathSync(script);
  // NOTE: `as any` is needed because @types/module-alias is at 2.0.0, while
  // module-alias is now at 2.2.0.
  (moduleAlias as any).addAlias(".", (fromPath: string) => {
    // A custom handler for local dependencies to handle cases where the entry
    // script is outside `/vcs`.

    // For entry scripts outside /vcs only, rewrite dot-slash-prefixed requires
    // to be rooted at `/vcs`.
    if (!fromPath.startsWith("/var/vcs") && fromPath === realScriptPath) {
      return "/var/vcs"
    }

    // For all other dot-slash-prefixed requires, resolve as usual.
    // NOTE: module-alias will not allow us to just return "." here, because
    // it uses path.join under the hood, which collapses "./foo" down to just
    // "foo", for which the module resolution semantics are different.  So,
    // return the directory of the requiring module, which gives the same result
    // as ".".
    return path.dirname(fromPath)
  })

  moduleAlias()
  require(script)
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

brigadier.fire(event, worker)

