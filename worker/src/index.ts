import * as fs from "fs"
import * as moduleAlias from "module-alias"
import * as path from "path"
import { run } from "./app"
import { ContextLogger, LogLevel } from "./brigadier/logger"

const logger = new ContextLogger([], LogLevel["LOG"])
const version = require("../package.json").version
logger.log(`brigade-worker version: ${version}`)

const scriptLocations = [
  "/var/vcs/brigade.js"
]

var script: string = ""
for (let scriptLocation of scriptLocations) {
  if (fs.existsSync(scriptLocation) && fs.readFileSync(scriptLocation, "utf8") != "") {
    script = scriptLocation
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
    if (!fromPath.startsWith("/vcs") && fromPath === realScriptPath) {
      return "/vcs"
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

run()
