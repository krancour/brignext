import * as kubernetes from "@kubernetes/client-node"
import * as fs from "fs"
import * as path from "path"

export const defaultClient = kubernetes.Config.defaultClient()

const retry = (fn, args, delay, times) => {
  // exponential back-off retry if status is in the 500s
  return fn.apply(defaultClient, args).catch(err => {
    if (
      times > 0 &&
      err.response &&
      500 <= err.response.statusCode &&
      err.response.statusCode < 600
    ) {
      return new Promise(resolve => {
        setTimeout(() => {
          resolve(retry(fn, args, delay * 2, times - 1))
        }, delay)
      })
    }
    return Promise.reject(err)
  })
}

const wrapClient = fns => {
  // wrap client methods with retry logic
  for (let fn of fns) {
    let originalFn = defaultClient[fn.name]
    defaultClient[fn.name] = function () {
      return retry(originalFn, arguments, 4000, 5)
    }
  }
}

wrapClient([
  defaultClient.readNamespacedSecret,
  defaultClient.createNamespacedConfigMap,
  defaultClient.createNamespacedPod,
  defaultClient.deleteNamespacedPod,
  defaultClient.readNamespacedPodLog
])

const getKubeConfig = (): kubernetes.KubeConfig => {
  const kc = new kubernetes.KubeConfig()
  const config =
    process.env.KUBECONFIG || path.join(process.env.HOME, ".kube", "config")
  if (fs.existsSync(config)) {
    kc.loadFromFile(config)
  } else {
    kc.loadFromCluster()
  }
  return kc
}

export const config = getKubeConfig()
