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

	"github.com/giantswarm/fleet-membership-operator-gcp/types"
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

func (c *Client) Register(ctx context.Context, cluster *capg.GCPCluster, oidcJwks []byte) (types.MembershipData, error) {
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
		return types.MembershipData{}, err
	}

	registeredMembership, err := op.Wait(ctx)
	if err != nil {
		logger.Error(err, "create membership operation failed")
		return types.MembershipData{}, err
	}
	return toMembershipData(registeredMembership), err
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
		logger.Error(err, "delete membership operation failed")
		return nil
	}

	return nil
}

func (c *Client) getMembership(ctx context.Context, cluster *capg.GCPCluster) (types.MembershipData, error) {
	req := &gkehubpb.GetMembershipRequest{
		Name: generateMembershipName(cluster),
	}

	membership, err := c.gkeClient.GetMembership(ctx, req)
	if err != nil {
		return types.MembershipData{}, err
	}

	return toMembershipData(membership), nil
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

func toMembershipData(membership *gkehubpb.Membership) types.MembershipData {
	return types.MembershipData{
		WorkloadIdentityPool: membership.Authority.WorkloadIdentityPool,
		IdentityProvider:     membership.Authority.IdentityProvider,
	}
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
