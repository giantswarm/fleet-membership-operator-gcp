package acceptance_test

import (
	"context"

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

	When("a cluster is created on a management cluster", func() {
		ctx = context.Background()

		clusterName = "acceptance-workload-cluster"

		BeforeEach(func() {
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

			Expect(ensureClusterCRExists(gcpCluster)).To(Succeed())
			patch := []byte(`{"status":{"ready":true}}`)

			Expect(k8sClient.Status().Patch(ctx, gcpCluster, client.RawPatch(types.MergePatchType, patch))).To(Succeed())
		})

		It("it should create a membership secret on the workload cluster", func() {
			membershipSecret := &corev1.Secret{}

			Eventually(func() error {
				err := workloadClient.Get(ctx, client.ObjectKey{
					Name:      controllers.MembershipSecretName,
					Namespace: controllers.DefaultMembershipSecretNamespace,
				}, membershipSecret)

				return err
			}, "120s").Should(Succeed())
		})
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
