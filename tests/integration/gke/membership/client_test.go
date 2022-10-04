package membership_test

import (
	"context"
	"fmt"
	"net/http"

	gkehub "cloud.google.com/go/gkehub/apiv1beta1"
	"cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/giantswarm/fleet-membership-operator-gcp/pkg/gke/membership"
	"github.com/giantswarm/fleet-membership-operator-gcp/tests"
	. "github.com/giantswarm/fleet-membership-operator-gcp/tests/matchers"
)

var _ = Describe("Client", func() {
	var (
		ctx context.Context

		gcpClient *gkehub.GkeHubMembershipClient
		client    *membership.Client

		jwks           []byte
		membershipName string
		cluster        *capg.GCPCluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger := zap.New(zap.WriteTo(GinkgoWriter))
		ctx = log.IntoContext(context.Background(), logger)

		jwks = generateJWKS()
		clusterName := tests.GenerateGUID("test")
		membershipName = fmt.Sprintf("projects/%s/locations/global/memberships/%s", gcpProject, clusterName)
		cluster = &capg.GCPCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
			},
		}

		var err error
		gcpClient, err = gkehub.NewGkeHubMembershipRESTClient(ctx)
		Expect(err).NotTo(HaveOccurred())

		client = membership.NewClient(gcpClient)
	})

	AfterEach(func() {
		req := &gkehubpb.DeleteMembershipRequest{
			Name: membershipName,
		}
		_, err := gcpClient.DeleteMembership(ctx, req)
		Expect(err).To(Or(
			Not(HaveOccurred()),
			BeGoogleAPIErrorWithStatus(http.StatusNotFound),
		))
	})

	It("creates the membership", func() {
		actualMembership, err := client.RegisterMembership(ctx, cluster, jwks)
		Expect(err).NotTo(HaveOccurred())

		getReq := &gkehubpb.GetMembershipRequest{
			Name: membershipName,
		}
		fleetMembership, err := gcpClient.GetMembership(ctx, getReq)
		Expect(err).NotTo(HaveOccurred())

		Expect(fleetMembership.Authority).NotTo(BeNil())
		Expect(fleetMembership.Authority.Issuer).To(Equal(membership.KubernetesIssuer))
		Expect(fleetMembership.Authority.OidcJwks).To(Equal(jwks))
		Expect(fleetMembership.Authority.IdentityProvider).NotTo(BeEmpty())
		Expect(fleetMembership.Authority.WorkloadIdentityPool).NotTo(BeEmpty())

		Expect(actualMembership).NotTo(BeNil())
		Expect(actualMembership.Authority).NotTo(BeNil())
		Expect(actualMembership.Authority.WorkloadIdentityPool).To(Equal(fleetMembership.Authority.WorkloadIdentityPool))
		Expect(actualMembership.Authority.IdentityProvider).To(Equal(fleetMembership.Authority.IdentityProvider))
	})

	When("the membership already exists", func() {
		BeforeEach(func() {
			_, err := client.RegisterMembership(ctx, cluster, jwks)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			actualMembership, err := client.RegisterMembership(ctx, cluster, jwks)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualMembership).NotTo(BeNil())
			Expect(actualMembership.Authority).NotTo(BeNil())
			Expect(actualMembership.Authority.IdentityProvider).NotTo(BeEmpty())
			Expect(actualMembership.Authority.WorkloadIdentityPool).NotTo(BeEmpty())
		})
	})

	When("the jwks are invalid", func() {
		BeforeEach(func() {
			jwks = []byte(`{"keys": [{"not-valid": true}]}`)
		})

		It("fails to register the membership", func() {
			actualMembership, err := client.RegisterMembership(ctx, cluster, jwks)
			Expect(err).To(HaveOccurred())
			Expect(actualMembership).To(BeNil())
		})
	})

	When("the client has insufficient premissions in the clusters project", func() {
		BeforeEach(func() {
			cluster.Spec.Project = "something-123"
		})

		It("fails to register the membership", func() {
			actualMembership, err := client.RegisterMembership(ctx, cluster, jwks)
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusForbidden))
			Expect(actualMembership).To(BeNil())
		})
	})
})
