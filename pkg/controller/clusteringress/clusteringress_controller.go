package clusteringress

import (
	"context"
	"fmt"
	"github.com/knative/pkg/kmeta"
	"github.com/knative/pkg/logging"
	"github.com/knative/serving/pkg/apis/networking"
	networkingv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
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

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterIngress{
		client: mgr.GetClient(),
		schema: mgr.GetScheme(),
	}
}

type ReconcileClusterIngress struct {
	client        client.Client
	schema        *runtime.Scheme
	kongNamespace string
}

func (r *ReconcileClusterIngress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	original := &networkingv1alpha1.ClusterIngress{}
	err := r.client.Get(ctx, request.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
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
	//TODO
	return reconcile.Result{}, nil
}

func (r *ReconcileClusterIngress) reconcile(ctx context.Context, ci *networkingv1alpha1.ClusterIngress) error {
	if ci.GetDeletionTimestamp() != nil {
		return r.reconcileDeletion(ctx, ci)
	}
	ci.SetDefaults(ctx)
	ci.Status.InitializeConditions()

	svc, err := MakeService(ctx, ci, r.kongNamespace)
	if err != nil {
		return err
	}
	if err := r.reconcileService(ctx, ci, svc); err != nil {
		return err
	}
	ci.Status.MarkNetworkConfigured()
	ci.Status.ObservedGeneration = ci.Generation
	return nil

}

func (r *ReconcileClusterIngress) reconcileDeletion(ctx context.Context, ci *networkingv1alpha1.ClusterIngress) error {
	return nil
}

func (r *ReconcileClusterIngress) reconcileService(ctx context.Context, ci *networkingv1alpha1.ClusterIngress,
	desired *corev1.Service) error {
	logger := logging.FromContext(ctx)
	svc := &corev1.Service{}
	err := r.client.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, svc)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(ctx, desired)
		if err != nil {
			logger.Errorw("Failed to create Ambassador config on K8s Service", err)
			return err
		}
		logger.Infof("Created Ambassador config on K8s Service %q in namespace %q", desired.Name, desired.Namespace)
	} else if err != nil {
		return err
	} else if !equality.Semantic.DeepEqual(svc.Spec, desired.Spec) || !equality.Semantic.DeepEqual(svc.ObjectMeta.Annotations, desired.ObjectMeta.Annotations) {
		// Don't modify the informers copy
		existing := svc.DeepCopy()
		existing.Spec = desired.Spec
		existing.ObjectMeta.Annotations = desired.ObjectMeta.Annotations
		err = r.client.Update(ctx, existing)
		if err != nil {
			logger.Errorw("Failed to update Ambassador config on K8s Service", err)
			return err
		}
	}
	return nil
}

func MakeService(ctx context.Context, ci *networkingv1alpha1.ClusterIngress, kongNamespace string) (*corev1.Service, error) {
	logger := logging.FromContext(ctx)
	ambassadorYaml := ""
	for _, rule := range ci.Spec.Rules {
		hosts := rule.Hosts
		logger.Info(hosts)
		for _, path := range rule.HTTP.Paths {
			prefix := path.Path
			logger.Info(prefix)
			for _, split := range path.Splits {
				service := fmt.Sprintf("%s.%s:%s", split.ServiceName, split.ServiceNamespace, split.ServicePort.String())
				logger.Info(service)
			}
		}
	}

	logger.Infof("Creating Ambassador Config:\n %s\n", ambassadorYaml)

	annotations := ci.ObjectMeta.Annotations
	annotations["getambassador.io/config"] = string(ambassadorYaml)

	labels := make(map[string]string)
	labels[networking.IngressLabelKey] = ci.Name

	ingressLabels := ci.Labels
	labels[serving.RouteLabelKey] = ingressLabels[serving.RouteLabelKey]
	labels[serving.RouteNamespaceLabelKey] = ingressLabels[serving.RouteNamespaceLabelKey]

	return &corev1.Service{
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
	}, nil
}
