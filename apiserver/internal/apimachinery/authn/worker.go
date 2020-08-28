package authn

import brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"

var workerPrincipal = &worker{}

type worker struct{}

func (w *worker) Roles() []brignext.Role {
	return nil
}
