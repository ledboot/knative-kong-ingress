package apis

import (
	networkingv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

var AddToSchemes runtime.SchemeBuilder

func init() {
	AddToSchemes = append(AddToSchemes, networkingv1alpha1.SchemeBuilder.AddToScheme)
}

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
