const process = require("process");
const fs = require("fs");
const { execFileSync } = require("child_process");
const path = require('path');

const event = require("/var/event/event.json");

let deps;
let configFilePath = path.join("/var/vcs", event.worker.configFilesDirectory, "brigade.json");
if (fs.existsSync(configFilePath)) {
  deps = require(configFilePath).dependencies || {};
} else {
  let configFileContents = event.worker.defaultConfigFiles["brigade.json"]
  if (configFileContents) {
    deps = JSON.parse(configFileContents).dependencies;
  }
}

if (require.main === module)  {
  addDeps()
}

function addDeps() {
  if (!deps) {
    console.log("prestart: no dependencies file found")
    return
  }

  const packages = buildPackageList(deps)
  if (packages.length == 0) {
    console.log("prestart: no dependencies to install")
    return
  }

  console.log(`prestart: installing ${packages.join(', ')}`)
  try {
    addYarn(packages)
  } catch (e)  {
    console.error(e)
    process.exit(1)
  }
}

function buildPackageList(deps) {
  if (!deps) {
    throw new Error("'deps' must not be null")
  }

  return Object.entries(deps).map(([dep, version]) => dep + "@" + version)
}

function addYarn(packages) {
  if (!packages || packages.length == 0) {
    throw new Error("'packages' must be an array with at least one item")
  }

  execFileSync("yarn", ["add", ...packages])
}
