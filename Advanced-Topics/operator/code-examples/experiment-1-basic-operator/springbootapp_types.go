package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SpringBootAppSpec defines the desired state of SpringBootApp
type SpringBootAppSpec struct {
    // Image is the container image for the Spring Boot application
    Image string `json:"image"`
    
    // Replicas is the number of desired replicas
    // +kubebuilder:default=1
    // +optional
    Replicas *int32 `json:"replicas,omitempty"`
    
    // Port is the port that the application listens on
    // +kubebuilder:default=8080
    // +optional
    Port int32 `json:"port,omitempty"`
}

// SpringBootAppStatus defines the observed state of SpringBootApp
type SpringBootAppStatus struct {
    // Replicas is the current number of replicas
    Replicas int32 `json:"replicas"`
    
    // ReadyReplicas is the number of ready replicas
    ReadyReplicas int32 `json:"readyReplicas"`
    
    // Phase represents the current phase of the application
    Phase string `json:"phase,omitempty"`
    
    // Conditions represent the latest available observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SpringBootApp is the Schema for the springbootapps API
type SpringBootApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   SpringBootAppSpec   `json:"spec,omitempty"`
    Status SpringBootAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpringBootAppList contains a list of SpringBootApp
type SpringBootAppList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []SpringBootApp `json:"items"`
}

func init() {
    SchemeBuilder.Register(&SpringBootApp{}, &SpringBootAppList{})
}