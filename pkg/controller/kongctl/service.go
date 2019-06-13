package kongctl

import (
	"github.com/hbagdi/go-kong/kong"
	"go.uber.org/zap"
)

const ControllerAgentName string = "kongctl.service.controller"

type KongController struct {
	client *kong.Client
	logger *zap.SugaredLogger
}

func NewKongController(client *kong.Client, logger *zap.SugaredLogger) *KongController {
	return &KongController{
		client: client,
		logger: logger,
	}
}

func (k *KongController) Get(nameOrId *string) (*kong.Service, error) {
	return k.client.Services.Get(nil, nameOrId)
}

func (c *KongController) Create(service *Service) {
	createService, err := c.client.Services.Create(nil, service.Service)
	if err != nil {
		c.logger.Errorf("create kongctl service error: %v", err)
	}

	for _, r := range service.Routes {
		r.Route.Service = createService
		_, err := c.client.Routes.Create(nil, r.Route)
		if err != nil {
			c.logger.Errorf("create kongctl route error: %v", err)
		}
	}
}

func (c *KongController) Update(service *Service) {
	updateService, err := c.client.Services.Update(nil, service.Service)
	if err != nil {
		c.logger.Errorf("update kongctl service error: %v", err)
		return
	}
	oldRoute, _, err := c.client.Routes.ListForService(nil, updateService.ID, nil)
	if err != nil {
		for _, r := range service.Routes {
			r.Route.Service = updateService
			if _, err := c.client.Routes.Create(nil, r.Route); err != nil {
				c.logger.Errorf("update kongctl route error: %v", err)
			}
		}
	} else {
		for _, or := range oldRoute {
			for _, nr := range service.Routes {
				if nr.Route.Name == or.Name {
					nr.Route.ID = or.ID
					c.client.Routes.Update(nil, nr.Route)
				}
			}
		}
	}
}

func (c *KongController) Delete(service *kong.Service) {
	routes, _, err := c.client.Routes.ListForService(nil, service.Name, nil)
	if err == nil {
		for _, r := range routes {
			if err := c.client.Routes.Delete(nil, r.Name); err != nil {
				c.logger.Errorf("delete kongctl route error: %v", err)
			}
		}
	}
	if err = c.client.Services.Delete(nil, service.Name); err != nil {
		c.logger.Errorf("delete kongctl service error: %v", err)
	}
	//delete route which is associated the service
	//for _, r := range service.Routes {
	//	if err := c.client.Routes.Delete(nil, r.Route.Name); err != nil {
	//		c.logger.Errorf("delete kongctl route error: %v", err)
	//	}
	//}
	//err := c.client.Services.Delete(nil, service.Service.Name)
	//if err != nil {
	//	c.logger.Errorf("delete kongctl service error: %v", err)
	//}
}
