package main

import (
	"fmt"

	providerconfig "github.com/kubermatic/machine-controller/pkg/providerconfig/types"
	dmzv1 "github.com/wozniakjan/cluster-wh/api"
	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	corev1 "k8s.io/api/core/v1"
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
	c.Spec.Cloud.Aws = kubermaticv0.AWSCloudSpec{
		CredentialsReference: *providerconfig.GlobalSecretKeySelector{
			ObjectReference: corev1.ObjectReference{
				Name:      "credentials-aws",
				Namespace: "jw-api-test",
			},
		},
	}
	cluster := kubermaticv1.Cluster{
		ObjectMeta: c.ObjectMeta,
		Spec:       c.Spec,
		Status: kubermaticv1.ClusterStatus{
			UserEmail: "jan@kubermatic.com",
		},
	}
	cluster.Name = ""
	cluster.GenerateName = fmt.Sprintf("%v-", c.Name)
	if err := s.kubermaticClient.Clusters().Create(ctx, cluster, metav1.CreateOptions{}); err != nil {
		return err
	}
	c.Spec = cluster.Spec
}
