package workload

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const KeyWorkloadClusterConfig = "value"

type ClusterConfigs struct {
	runtimeClient client.Client
}

func NewClusterConfigs(client client.Client) *ClusterConfigs {
	return &ClusterConfigs{
		runtimeClient: client,
	}
}

func (c *ClusterConfigs) Get(ctx context.Context, cluster *capg.GCPCluster) (*rest.Config, error) {
	logger := c.getLogger(ctx)

	secret := &corev1.Secret{}
	secretName := fmt.Sprintf("%s-kubeconfig", cluster.Name)
	logger = logger.WithValues("secret-name", secretName)

	err := c.runtimeClient.Get(ctx, k8stypes.NamespacedName{
		Name:      secretName,
		Namespace: cluster.Namespace,
	}, secret)
	if err != nil {
		logger.Error(err, "could not get cluster secret")
		return nil, err
	}

	data, ok := secret.Data[KeyWorkloadClusterConfig]
	if !ok {
		err = errors.New("cluster kubeconfig data is missing")
		logger.Error(err, "failed to get kubeconfig")
		return nil, err
	}

	config, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		logger.Error(err, "failed to build client from kubeconfig")
		return nil, err
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to get client config")
		return nil, err
	}

	return clientConfig, nil
}

func (c *ClusterConfigs) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("workload-clients")
}
