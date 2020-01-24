package kubernetes

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const envconfigPrefix = "KUBE"

// config represents common configuration options for a Kubernetes client
type config struct {
	MasterURL      string `envconfig:"MASTER"`
	KubeConfigPath string `envconfig:"CONFIG"`
	Namespace      string `envconfig:"BRIGADE_NAMESPACE" default:"brigade"`
}

// Client returns a new Kubernetes client
func Client() (*kubernetes.Clientset, error) {
	c := config{}
	if err := envconfig.Process(envconfigPrefix, &c); err != nil {
		return nil, errors.Wrap(
			err,
			"error getting kubernetes configuration from environment",
		)
	}
	var cfg *rest.Config
	var err error
	if c.MasterURL == "" && c.KubeConfigPath == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(c.MasterURL, c.KubeConfigPath)
	}
	if err != nil {
		return nil, errors.Wrap(
			err,
			"error getting kubernetes configuration",
		)
	}
	return kubernetes.NewForConfig(cfg)
}

// BrigadeNamespace returns the Kubernetes namespace used by Brigade
func BrigadeNamespace() (string, error) {
	c := config{}
	if err := envconfig.Process(envconfigPrefix, &c); err != nil {
		return "", errors.Wrap(
			err,
			"error getting kubernetes configuration from environment",
		)
	}
	return c.Namespace, nil
}
