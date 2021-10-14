package main

import (
	"context"
	"fmt"

	providerconfig "github.com/kubermatic/machine-controller/pkg/providerconfig/types"
	dmzv1 "github.com/wozniakjan/cluster-wh/api"
	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *server) proxyCluster(c *dmzv1.DMZCluster) error {
	// hardcodings
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Annotations["kubermatic.io/aws-region"] = "eu-central-1"
	c.Labels["project-id"] = "testjw"
	c.Spec.ExposeStrategy = kubermaticv1.ExposeStrategyTunneling
	c.Spec.HumanReadableName = c.Name
	c.Spec.Cloud.AWS = &kubermaticv1.AWSCloudSpec{
		CredentialsReference: &providerconfig.GlobalSecretKeySelector{
			ObjectReference: corev1.ObjectReference{
				Name:      "credentials-aws",
				Namespace: "jw-api-test",
			},
		},
	}
	cluster := &kubermaticv1.Cluster{
		ObjectMeta: c.ObjectMeta,
		Spec:       c.Spec.ClusterSpec,
		Status: kubermaticv1.ClusterStatus{
			UserEmail: "jan@kubermatic.com",
		},
	}
	cluster.Name = ""
	cluster.GenerateName = fmt.Sprintf("%v-", c.Name)
	cluster2, err := s.kubermaticClient.Clusters().Create(context.TODO(), cluster, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	c.Spec.ClusterSpec = cluster2.Spec
	return nil
}
