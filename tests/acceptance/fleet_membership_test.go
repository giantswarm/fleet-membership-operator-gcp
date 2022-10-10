package acceptance_test

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/fleet-membership-operator-gcp/controllers"
)

var _ = Describe("Fleet Membership", func() {
	var (
		ctx context.Context

		gcpCluster  *capg.GCPCluster
		clusterName string
	)

	BeforeEach(func() {
		ctx = context.Background()

		clusterName = "acceptance-workload-cluster"
		gcpCluster = &capg.GCPCluster{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: "giantswarm",
				Annotations: map[string]string{
					controllers.AnnotationWorkloadIdentityEnabled: "true",
				},
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
			},
		}

		err := k8sClient.Create(context.Background(), gcpCluster)
		if !k8serrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		patch := []byte(`{"status":{"ready":true}}`)

		Expect(k8sClient.Status().Patch(ctx, gcpCluster, client.RawPatch(types.MergePatchType, patch))).To(Succeed())
	})

	AfterEach(func() {
		membershipSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      controllers.MembershipSecretName,
				Namespace: controllers.DefaultMembershipSecretNamespace,
			},
		}
		err := workloadClient.Delete(ctx, membershipSecret)

		Expect(err).NotTo(HaveOccurred())
	})

	It("reconciles the cluster", func() {
		By("creating a membership secret on the workload cluster")

		membershipSecret := &corev1.Secret{}
		Eventually(func() error {
			err := workloadClient.Get(ctx, client.ObjectKey{
				Name:      controllers.MembershipSecretName,
				Namespace: controllers.DefaultMembershipSecretNamespace,
			}, membershipSecret)

			return err
		}, "120s").Should(Succeed())

		data := membershipSecret.Data[controllers.SecretKeyGoogleApplicationCredentials]
		var actualMembership gkehubpb.Membership
		Expect(json.Unmarshal(data, &actualMembership)).To(Succeed())

		By("not preventing cluster deletion")

		Expect(k8sClient.Delete(ctx, gcpCluster)).To(Succeed())
		Eventually(func(g Gomega) bool {
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      gcpCluster.Name,
				Namespace: gcpCluster.Namespace,
			}, &capg.GCPCluster{})
			if !k8serrors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
				return false
			}
			return true
		}, "120s").Should(BeTrue())
	})
})

func ensureClusterCRExists(gcpCluster *capg.GCPCluster) error {
	ctx := context.Background()

	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      gcpCluster.Name,
		Namespace: gcpCluster.Namespace,
	}, gcpCluster)

	if k8serrors.IsNotFound(err) {
		err = k8sClient.Create(context.Background(), gcpCluster)
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return err
}
