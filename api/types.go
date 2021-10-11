package v1

import (
	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterSpec defines the cluster specification
type ClusterSpec struct {
	// Cloud specifies the cloud providers configuration
	Cloud kubermaticv1.CloudSpec `json:"cloud"`

	// MachineNetworks optionally specifies the parameters for IPAM.
	MachineNetworks []kubermaticv1.MachineNetworkingConfig `json:"machineNetworks,omitempty"`

	// OIDC settings
	OIDC kubermaticv1.OIDCSettings `json:"oidc,omitempty"`

	// Configure cluster upgrade window, currently used for flatcar node reboots
	UpdateWindow *kubermaticv1.UpdateWindow `json:"updateWindow,omitempty"`

	// If active the PodSecurityPolicy admission plugin is configured at the apiserver
	UsePodSecurityPolicyAdmissionPlugin bool `json:"usePodSecurityPolicyAdmissionPlugin,omitempty"`

	// If active the PodNodeSelector admission plugin is configured at the apiserver
	UsePodNodeSelectorAdmissionPlugin bool `json:"usePodNodeSelectorAdmissionPlugin,omitempty"`

	// EnableUserSSHKeyAgent control whether the UserSSHKeyAgent will be deployed in the user cluster or not.
	// If it was enabled, the agent will be deployed and used to sync the user ssh keys, that the user attach
	// to the created cluster. If the agent was disabled, it won't be deployed in the user cluster, thus after
	// the cluster creation any attached ssh keys won't be synced to the worker nodes. Once the agent is enabled/disabled
	// it cannot be changed after the cluster is being created.
	EnableUserSSHKeyAgent *bool `json:"enableUserSSHKeyAgent,omitempty"`

	// PodNodeSelectorAdmissionPluginConfig provides the configuration for the PodNodeSelector.
	// It's used by the backend to create a configuration file for this plugin.
	// The key:value from the map is converted to the namespace:<node-selectors-labels> in the file.
	// The format in a file:
	// podNodeSelectorPluginConfig:
	//  clusterDefaultNodeSelector: <node-selectors-labels>
	//  namespace1: <node-selectors-labels>
	//  namespace2: <node-selectors-labels>
	PodNodeSelectorAdmissionPluginConfig map[string]string `json:"podNodeSelectorAdmissionPluginConfig,omitempty"`

	// Additional Admission Controller plugins
	AdmissionPlugins []string `json:"admissionPlugins,omitempty"`

	// AuditLogging
	AuditLogging *kubermaticv1.AuditLoggingSettings `json:"auditLogging,omitempty"`

	// ServiceAccount contains service account related settings for the kube-apiserver of user cluster.
	ServiceAccount *kubermaticv1.ServiceAccountSettings `json:"serviceAccount,omitempty"`

	// OPAIntegration is a preview feature that enables OPA integration with Kubermatic for the cluster.
	// Enabling it causes gatekeeper and its resources to be deployed on the user cluster.
	// By default it is disabled.
	OPAIntegration *kubermaticv1.OPAIntegrationSettings `json:"opaIntegration,omitempty"`

	// MLA contains monitoring, logging and alerting related settings for the user cluster.
	MLA *kubermaticv1.MLASettings `json:"mla,omitempty"`

	// ClusterNetwork contains network settings.
	ClusterNetwork *kubermaticv1.ClusterNetworkingConfig `json:"clusterNetwork,omitempty"`

	// ContainerRuntime to use, i.e. Docker or containerd. By default containerd will be used.
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// ClusterStatus defines the observed state
type ClusterStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.cloud.dc", name="DataCenter", type="string"
// +kubebuilder:printcolumn:JSONPath=".spec.version", name="Version", type="string"
// +kubebuilder:printcolumn:JSONPath=".spec.containerRuntime", name="CRI", type="string", priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.cniPlugin.type", name="CNI", type="string", priority=1
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name="Age",type="date",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// Cluster is the Schema for the cluster DMZ API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterList contains a list of Clusters
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
