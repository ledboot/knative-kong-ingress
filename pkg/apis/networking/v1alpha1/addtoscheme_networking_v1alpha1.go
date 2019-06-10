package v1alpha1

import (
	networkingv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"github.com/ledboot/knative-kong-ingress/pkg/apis"
)

func init() {
	apis.AddToSchemes = append(apis.AddToSchemes, networkingv1alpha1.SchemeBuilder.AddToScheme)
}