package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type DeploymentTemplate struct {
	Path string `json:"path"`
	Parameters map[string]string `json:"parameters"`
}

type StatusPhase string

var (
	NoPhase        StatusPhase = ""
	ReadyPhase     StatusPhase = "Ready"
	ProvisionPhase StatusPhase = "Provisioning"
	ErrorPhase     StatusPhase = "Error"
)

// TDeploymentSpec defines the desired state of TDeployment
type TDeploymentSpec struct {
	Template DeploymentTemplate `json:"template"`
}

// TDeploymentStatus defines the observed state of TDeployment
type TDeploymentStatus struct {
	Phase StatusPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TDeployment is the Schema for the deployments API
// +k8s:openapi-gen=true
type TDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TDeploymentSpec   `json:"spec,omitempty"`
	Status TDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TDeploymentList contains a list of TDeployment
type TDeploymentList struct {
	metav1.TypeMeta               `json:",inline"`
	metav1.ListMeta               `json:"metadata,omitempty"`
	Items           []TDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TDeployment{}, &TDeploymentList{})
}
