package apis

import (
	"github.com/integr8ly/deployment-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/schemes"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, schemes.AddToScheme)
}
