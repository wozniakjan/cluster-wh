package api

import (
	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterSpec defines the cluster specification
type ClusterSpec struct {
	// kubermatic spec
	kubermaticv1.ClusterSpec `json:",inline"`
}

// ClusterStatus defines the observed state
type ClusterStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.cloud.dc", name="DataCenter", type="string"
// +kubebuilder:printcolumn:JSONPath=".spec.version", name="Version", type="string"
// +kubebuilder:printcolumn:JSONPath=".spec.containerRuntime", name="CRI", type="string", priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.cniPlugin.type", name="CNI", type="string", priority=1
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name="Age",type="date",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// DMZCluster is the Schema for the cluster DMZ API
type DMZCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// DMZClusterList contains a list of Clusters
type DMZClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DMZCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DMZCluster{}, &DMZClusterList{})
}
