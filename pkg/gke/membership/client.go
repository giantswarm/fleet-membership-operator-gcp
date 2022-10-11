package membership

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	gkehub "cloud.google.com/go/gkehub/apiv1beta1"
	gkehubpb "cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	"github.com/go-logr/logr"
	"google.golang.org/api/googleapi"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KubernetesIssuer = "https://kubernetes.default.svc.cluster.local"
)

type Client struct {
	gkeClient *gkehub.GkeHubMembershipClient
}

func NewClient(client *gkehub.GkeHubMembershipClient) *Client {
	return &Client{
		gkeClient: client,
	}
}

func (c *Client) Register(ctx context.Context, cluster *capg.GCPCluster, oidcJwks []byte) (*gkehubpb.Membership, error) {
	logger := c.getLogger(ctx)
	logger.Info("registering membership")
	defer logger.Info("done registering membership")

	membership := generateMembership(cluster, oidcJwks)
	parent := fmt.Sprintf("projects/%s/locations/global", cluster.Spec.Project)
	req := &gkehubpb.CreateMembershipRequest{
		Parent:       parent,
		MembershipId: cluster.Name,
		Resource:     membership,
	}

	op, err := c.gkeClient.CreateMembership(ctx, req)
	if hasHttpCode(err, http.StatusConflict) {
		logger.Info(fmt.Sprintf("membership %s already exists", membership.Name))
		return c.getMembership(ctx, cluster)
	}
	if err != nil {
		logger.Error(err, "failed to create membership")
		return nil, err
	}

	registeredMembership, err := op.Wait(ctx)
	if err != nil {
		logger.Error(err, "create membership operation failed")
		return nil, err
	}
	return registeredMembership, err
}

func (c *Client) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := c.getLogger(ctx)
	logger.Info("unregistering membership")
	defer logger.Info("done unregistering membership")

	req := &gkehubpb.DeleteMembershipRequest{
		Name: generateMembershipName(cluster),
	}
	op, err := c.gkeClient.DeleteMembership(ctx, req)
	if hasHttpCode(err, http.StatusNotFound) {
		logger.Info("membership already unregistered")
		return nil
	}
	if err != nil {
		logger.Error(err, "failed to delete membership")
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		logger.Error(err, "create membership operation failed")
		return nil
	}

	return nil
}

func (c *Client) getMembership(ctx context.Context, cluster *capg.GCPCluster) (*gkehubpb.Membership, error) {
	req := &gkehubpb.GetMembershipRequest{
		Name: generateMembershipName(cluster),
	}

	return c.gkeClient.GetMembership(ctx, req)
}

func (c *Client) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("gke-client")
}

func generateMembership(cluster *capg.GCPCluster, oidcJwks []byte) *gkehubpb.Membership {
	name := generateMembershipName(cluster)

	membership := &gkehubpb.Membership{
		Name: name,
		Authority: &gkehubpb.Authority{
			Issuer:   KubernetesIssuer,
			OidcJwks: oidcJwks,
		},
	}

	return membership
}

func generateMembershipName(cluster *capg.GCPCluster) string {
	return fmt.Sprintf("projects/%s/locations/global/memberships/%s", cluster.Spec.Project, cluster.Name)
}

func hasHttpCode(err error, statusCode int) bool {
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		if googleErr.Code == statusCode {
			return true
		}
	}

	return false
}
