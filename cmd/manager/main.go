package main

import (
	"github.com/ledboot/knative-kong-ingress/pkg/apis"
	"github.com/ledboot/knative-kong-ingress/pkg/controller"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logf.Log.WithName("cmd")

func main() {

	logf.SetLogger(zap.Logger())
	cfg, err := config.GetConfig()

	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	//ctx := context.TODO()

	mgr, err := manager.New(cfg, manager.Options{})

	log.Info("api config : %v", mgr.GetConfig())
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}

}
