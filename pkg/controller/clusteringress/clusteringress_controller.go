package clusteringress

import (
	"context"
	"fmt"
	"github.com/hbagdi/go-kong/kong"
	"github.com/knative/pkg/kmeta"
	"github.com/knative/pkg/logging"
	"github.com/knative/serving/pkg/apis/networking"
	networkingv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/network"
	"github.com/ledboot/knative-kong-ingress/pkg/controller/kongctl"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	kongIngressClass = "kong.ingress.networking.knative.dev"
)

func Add(mgr manager.Manager, kongController *kongctl.KongController) error {
	return add(mgr, newReconciler(mgr, kongController))
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("clusteringress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &networkingv1alpha1.ClusterIngress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &networkingv1alpha1.ClusterIngress{},
	})
	if err != nil {
		return err
	}
	return nil
}

func newReconciler(mgr manager.Manager, kongController *kongctl.KongController) reconcile.Reconciler {
	return &ReconcileClusterIngress{
		client:         mgr.GetClient(),
		schema:         mgr.GetScheme(),
		kongController: kongController,
		kongNamespace:  "kong",
	}

}

type ReconcileClusterIngress struct {
	client         client.Client
	schema         *runtime.Scheme
	kongController *kongctl.KongController
	kongNamespace  string
}

func (r *ReconcileClusterIngress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	logger := logging.FromContext(ctx)
	original := &networkingv1alpha1.ClusterIngress{}
	err := r.client.Get(ctx, request.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// TODO can't get delete object info
			//r.reconcileDeletion(ctx, original)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	ingressClass := original.ObjectMeta.Annotations[networking.IngressClassAnnotationKey]
	if ingressClass != kongIngressClass {
		return reconcile.Result{}, nil
	}
	ci := original.DeepCopy()
	err = r.reconcile(ctx, ci)

	if equality.Semantic.DeepEqual(original.Status, ci.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err := r.updateStatus(ctx, ci); err != nil {
		logger.Warnw("Failed to update clusterIngress status", err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileClusterIngress) updateStatus(ctx context.Context, desired *networkingv1alpha1.ClusterIngress) (*networkingv1alpha1.ClusterIngress, error) {
	ci := &networkingv1alpha1.ClusterIngress{}
	err := r.client.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, ci)
	if err != nil {
		return nil, err
	}

	// If there's nothing to update, just return.
	if reflect.DeepEqual(ci.Status, desired.Status) {
		return ci, nil
	}
	// Don't modify the informers copy
	existing := ci.DeepCopy()
	existing.Status = desired.Status
	err = r.client.Status().Update(ctx, existing)
	return existing, err
}

func (r *ReconcileClusterIngress) reconcile(ctx context.Context, ci *networkingv1alpha1.ClusterIngress) error {
	if ci.GetDeletionTimestamp() != nil {
		return r.reconcileDeletion(ctx, ci)
	}
	ci.SetDefaults(ctx)
	ci.Status.InitializeConditions()

	kongState, err := MakeService(ctx, ci, r.kongNamespace)
	if err != nil {
		return err
	}
	if err := r.reconcileService(ctx, ci, kongState); err != nil {
		return err
	}
	ci.Status.MarkNetworkConfigured()
	ci.Status.ObservedGeneration = ci.Generation
	ci.Status.MarkLoadBalancerReady(getLBStatus(r.kongNamespace))
	return nil

}

func (r *ReconcileClusterIngress) reconcileDeletion(ctx context.Context, ci *networkingv1alpha1.ClusterIngress) error {
	status := makeKongService(ci)
	if len(status.Services) < 1 {
		return fmt.Errorf("no service to delete!")
	}
	if service, err := r.kongController.Get(status.Services[0].Service.Name); err != nil {
		r.kongController.Delete(service)
	}
	return nil
}

func (r *ReconcileClusterIngress) reconcileService(ctx context.Context, ci *networkingv1alpha1.ClusterIngress, desired *kongctl.KongState) error {
	logger := logging.FromContext(ctx)
	svc := &corev1.Service{}
	err := r.client.Get(ctx, types.NamespacedName{Name: desired.CoreService.Name, Namespace: desired.CoreService.Namespace}, svc)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(ctx, desired.CoreService)

		if err != nil {
			logger.Errorw("Failed to create kong config on K8s Service", err)
			return err
		}
		//create kongctl service
		for _, svc := range desired.Services {
			r.kongController.Create(svc)
		}

		logger.Infof("Created kong config on K8s Service %q in namespace %q", desired.CoreService.Name, desired.CoreService.Namespace)
	} else if err != nil {
		return err
	} else if !equality.Semantic.DeepEqual(svc.Spec, desired.CoreService.Spec) || !equality.Semantic.DeepEqual(svc.ObjectMeta.Annotations, desired.CoreService.ObjectMeta.Annotations) {
		// Don't modify the informers copy
		existing := svc.DeepCopy()
		existing.Spec = desired.CoreService.Spec
		existing.ObjectMeta.Annotations = desired.CoreService.ObjectMeta.Annotations
		err = r.client.Update(ctx, existing)
		if err != nil {
			logger.Errorw("Failed to update kong config on K8s Service", err)
			return err
		}
		for _, desired_svc := range desired.Services {
			if service, err := r.kongController.Get(desired_svc.Service.Name); err == nil {
				desired_svc.Service.ID = service.ID
				r.kongController.Update(desired_svc)
			}
		}

	}
	return nil
}

func makeKongService(ci *networkingv1alpha1.ClusterIngress) *kongctl.KongState {
	state := &kongctl.KongState{}
	for _, rule := range ci.Spec.Rules {
		hosts := rule.Hosts
		for _, path := range rule.HTTP.Paths {
			for _, split := range path.Splits {

				kongHost := fmt.Sprintf("%s.%s.%s", split.ServiceName, split.ServiceNamespace, "svc.cluster.local")
				serviceName := fmt.Sprintf("%s-%s", split.ServiceName, split.ServiceNamespace)
				routeName := fmt.Sprintf("%s-%s", serviceName, "route")

				kongService := &kong.Service{
					Host: kong.String(kongHost),
					Name: kong.String(serviceName),
					Port: kong.Int(split.ServicePort.IntValue()),
				}
				kongRoute := &kong.Route{
					Name:      kong.String(routeName),
					Hosts:     kong.StringSlice(hosts...),
					Protocols: kong.StringSlice("http", "https"),
				}
				route := &kongctl.Route{
					Route: kongRoute,
				}
				service := &kongctl.Service{
					Service: kongService,
				}
				service.Routes = append(service.Routes, route)
				state.Services = append(state.Services, service)
			}
		}
	}
	return state
}

func MakeService(ctx context.Context, ci *networkingv1alpha1.ClusterIngress, kongNamespace string) (*kongctl.KongState, error) {
	state := makeKongService(ci)
	annotations := ci.ObjectMeta.Annotations

	labels := make(map[string]string)
	labels[networking.IngressLabelKey] = ci.Name

	ingressLabels := ci.Labels
	labels[serving.RouteLabelKey] = ingressLabels[serving.RouteLabelKey]
	labels[serving.RouteNamespaceLabelKey] = ingressLabels[serving.RouteNamespaceLabelKey]

	coreSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ci.Name,
			Namespace:       kongNamespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(ci)},
			Labels:          labels,
			Annotations:     annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "None",
		},
	}
	state.CoreService = coreSvc
	return state, nil
}

func getLBStatus(kongNamespace string) []networkingv1alpha1.LoadBalancerIngressStatus {
	// TODO: something better...
	return []networkingv1alpha1.LoadBalancerIngressStatus{
		{DomainInternal: fmt.Sprintf("kong-proxy.%s.svc.%s", kongNamespace, network.GetClusterDomainName())},
	}
}
