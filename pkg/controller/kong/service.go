package kong

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/ledboot/knative-kong-ingress/pkg/controller/clusteringress"
)

const ControllerAgentName string = "kong.service.controller"



func (r clusteringress.ReconcileClusterIngress) Get(nameOrId *string) (*kong.Service, error) {
	return c.client.Services.Get(nil, nameOrId)
}

func (r *ReconcileClusterIngress) Create(service Service) {
	createService, err := c.client.Services.Create(nil, &service.Service)
	if err != nil {
		c.log.Errorf("create kong service error: %v", err)
	}

	for _, r := range service.Routes {
		r.Route.Service = createService
		_, err := c.client.Routes.Create(nil, &r.Route)
		if err != nil {
			c.log.Errorf("create kong route error: %v", err)
		}
	}
}

func (r *ReconcileClusterIngress) Update(service Service) {
	updateService, err := c.client.Services.Update(nil, &service.Service)
	if err != nil {
		c.log.Errorf("update kong service error: %v", err)
		return
	}
	oldRoute, _, err := c.client.Routes.ListForService(nil, updateService.ID, nil)
	if err != nil {
		for _, r := range service.Routes {
			r.Route.Service = updateService
			if _, err := c.client.Routes.Create(nil, &r.Route); err != nil {
				c.log.Errorf("update kong route error: %v", err)
			}
		}
	} else {
		for _, or := range oldRoute {
			for _, nr := range service.Routes {
				if nr.Route.Name == or.Name {
					nr.Route.ID = or.ID
					c.client.Routes.Update(nil, &nr.Route)
				}
			}
		}
	}
}

func (r *ReconcileClusterIngress) Delete(service Service) {
	//delete route which is associated the service
	for _, r := range service.Routes {
		if err := c.client.Routes.Delete(nil, r.Name); err != nil {
			c.log.Errorf("delete kong route error: %v", err)
		}
	}
	err := c.client.Services.Delete(nil, service.Name)
	if err != nil {
		c.log.Errorf("delete kong service error: %v", err)
	}
}
