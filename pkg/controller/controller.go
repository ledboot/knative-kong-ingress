package controller

import (
	"github.com/hbagdi/go-kong/kong"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var AddToManagerFuncs []func(manager.Manager, *kong.Client) error

func AddToManager(m manager.Manager, client *kong.Client) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, client); err != nil {
			return err
		}
	}
	return nil
}