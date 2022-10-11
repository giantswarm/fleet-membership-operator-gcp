package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/fleet-membership-operator-gcp/types"
)

const (
	SecretKeyGoogleApplicationCredentials = "config"
	MembershipSecretName                  = "fleet-membership-operator-gcp-membership"
	DefaultMembershipDataNamespace        = "giantswarm"
)

type Cluster struct {
	config *rest.Config
	client client.Client
}

func NewCluster(config *rest.Config) (*Cluster, error) {
	runtimeClient, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return &Cluster{
		config: config,
		client: runtimeClient,
	}, nil
}

func (c *Cluster) GetOIDCJWKS() ([]byte, error) {
	reqUrl := fmt.Sprintf("%s/openid/v1/jwks", c.config.Host)

	httpClient, err := rest.HTTPClientFor(c.config)
	if err != nil {
		return []byte{}, err
	}

	resp, err := httpClient.Get(reqUrl)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func (c *Cluster) SaveMembershipData(ctx context.Context, namespace string, membership types.MembershipData) error {
	membershipJson, err := json.Marshal(membership)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MembershipSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			SecretKeyGoogleApplicationCredentials: string(membershipJson),
		},
	}

	err = c.client.Create(ctx, secret)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
