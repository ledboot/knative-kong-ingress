package main

import (
	"flag"
	"github.com/hbagdi/go-kong/kong"
	"github.com/ledboot/knative-kong-ingress/pkg/apis"
	"github.com/ledboot/knative-kong-ingress/pkg/configmap"
	"github.com/ledboot/knative-kong-ingress/pkg/controller"
	"github.com/ledboot/knative-kong-ingress/pkg/controller/kongctl"
	"github.com/ledboot/knative-kong-ingress/pkg/logging"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const component = "cmd"

var (
	kongAdminURL = flag.String("kong_admin_url", "http://192.168.64.5:32216", "Kong Admin URL")
)

func main() {

	//loggingConfigMap, err := configmap.Load("/etc/config-logging")
	loggingConfigMap, err := configmap.Load("/Users/gwynn/go/src/github.com/ledboot/knative-kong-ingress/config-logging")
	if err != nil {
		log.Fatalf("Error loading logging configuration: %v", err)
	}
	loggingConfig, err := logging.NewConfigFromMap(loggingConfigMap)
	if err != nil {
		log.Fatalf("Error parsing logging configuration: %v", err)
	}
	logger, _ := logging.NewLogger(loggingConfig, component)
	defer logger.Sync()

	cfg, err := config.GetConfig()

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	mgr, err := manager.New(cfg, manager.Options{})

	logger.Info(mgr.GetConfig().Host)

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	kongClient, err := kong.NewClient(kongAdminURL, nil)
	if err != nil {
		logger.Error("make kongctl client error :", err)
		os.Exit(1)
	}
	root, err := kongClient.Root(nil)
	if err != nil {
		logger.Error(err, "can not connect kongctl admin")
		os.Exit(1)
	}

	kongController := kongctl.NewKongController(kongClient, logger)

	logger.Infof("kongctl version : %s", root["version"])

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	if err := controller.AddToManager(mgr, kongController); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	logger.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}

}
