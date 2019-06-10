package main

import (
	"context"
	"flag"
	"github.com/hbagdi/go-kong/kong"
	"github.com/knative/pkg/logging"
	"github.com/ledboot/knative-kong-ingress/pkg/apis"
	"github.com/ledboot/knative-kong-ingress/pkg/controller"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	kongAdminURL = flag.String("kong_admin_url", "http://192.168.64.5:32216", "Kong Admin URL")
)

func main() {
	ctx := context.TODO()
	logf.SetLogger(zap.Logger())
	log := logging.FromContext(ctx)
	cfg, err := config.GetConfig()

	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	mgr, err := manager.New(cfg, manager.Options{})

	log.Info(mgr.GetConfig().Host)

	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	kongClient, err := kong.NewClient(kongAdminURL, nil)
	if err != nil {
		log.Error("make kong client error :", err)
		os.Exit(1)
	}
	root, err := kongClient.Root(nil)
	if err != nil {
		log.Error(err, "can not connect kong admin")
		os.Exit(1)
	}

	log.Infof("kong version : %s", root["version"])

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := controller.AddToManager(mgr,kongClient); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}

}
