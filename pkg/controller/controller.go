package controller

import (
	"github.com/ledboot/knative-kong-ingress/pkg/controller/clusteringress"
	"github.com/ledboot/knative-kong-ingress/pkg/controller/kongctl"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusteringress.Add)
}

var AddToManagerFuncs []func(manager.Manager, *kongctl.KongController) error

func AddToManager(m manager.Manager, client *kongctl.KongController) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, client); err != nil {
			return err
		}
	}
	return nil
}
