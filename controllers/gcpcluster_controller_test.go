package controllers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gkehubpb "cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/fleet-membership-operator-gcp/controllers"
	"github.com/giantswarm/fleet-membership-operator-gcp/controllers/controllersfakes"
	"github.com/giantswarm/fleet-membership-operator-gcp/pkg/gke/membership"
	"github.com/giantswarm/fleet-membership-operator-gcp/tests"
)

var _ = Describe("GCPCluster Reconcilation", func() {
	const (
		clusterName = "krillin"
		gcpProject  = "testing-1234"
		timeout     = time.Second * 5
		interval    = time.Millisecond * 250
	)

	var (
		ctx context.Context

		fakeGKEClient     *controllersfakes.FakeGKEMembershipClient
		clusterReconciler *controllers.GCPClusterReconciler

		gcpCluster          *capg.GCPCluster
		kubeadmControlPlane *capi.KubeadmControlPlane
		kubeconfigSecret    *corev1.Secret

		result      reconcile.Result
		reconcilErr error
	)

	BeforeEach(func() {
		SetDefaultConsistentlyDuration(timeout)
		SetDefaultConsistentlyPollingInterval(interval)
		SetDefaultEventuallyPollingInterval(interval)
		SetDefaultEventuallyTimeout(timeout)
		ctx = context.Background()

		gcpCluster = &capg.GCPCluster{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
				Annotations: map[string]string{
					controllers.AnnotationWorkloadIdentityEnabled: "true",
				},
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
			},
			Status: capg.GCPClusterStatus{
				Ready: true,
			},
		}

		kubeadmControlPlane = &capi.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
		}
		Expect(k8sClient.Create(ctx, kubeadmControlPlane)).To(Succeed())

		controlPlaneStatus := capi.KubeadmControlPlaneStatus{
			Ready: true,
		}

		tests.PatchControlPlaneStatus(k8sClient, kubeadmControlPlane, controlPlaneStatus)

		Expect(k8sClient.Create(ctx, gcpCluster)).To(Succeed())
		clusterStatus := capg.GCPClusterStatus{
			Ready: true,
		}
		tests.PatchClusterStatus(k8sClient, gcpCluster, clusterStatus)

		secretName := fmt.Sprintf("%s-kubeconfig", gcpCluster.Name)
		kubeconfig, err := KubeConfigFromREST(cfg)

		Expect(err).To(BeNil())

		kubeconfigSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"value": kubeconfig,
			},
		}
		Expect(k8sClient.Create(ctx, kubeconfigSecret)).To(Succeed())

		fakeMembership := &gkehubpb.Membership{
			Name: "/project/the-project/locations/global/membership/the-membership",
			Authority: &gkehubpb.Authority{
				Issuer:               membership.KubernetesIssuer,
				WorkloadIdentityPool: "the-workload-id-pool",
				IdentityProvider:     "the-identity-provider",
				OidcJwks:             []byte("the jwks"),
			},
		}
		fakeGKEClient = new(controllersfakes.FakeGKEMembershipClient)
		fakeGKEClient.RegisterReturns(fakeMembership, nil)

		clusterReconciler = &controllers.GCPClusterReconciler{
			Client:                    k8sClient,
			Logger:                    logf.Log,
			MembershipSecretNamespace: namespace,
			GKEMembershipClient:       fakeGKEClient,
		}
	})

	JustBeforeEach(func() {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      gcpCluster.Name,
				Namespace: gcpCluster.Namespace,
			},
		}
		result, reconcilErr = clusterReconciler.Reconcile(ctx, req)
	})

	It("reconciles successfully", func() {
		Expect(reconcilErr).NotTo(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())
	})

	It("creates a gke membership secret with the correct credentials", func() {
		secret := &corev1.Secret{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      controllers.MembershipSecretName,
			Namespace: namespace,
		}, secret)
		Expect(err).NotTo(HaveOccurred())

		Expect(secret).ToNot(BeNil())
		Expect(secret.Annotations).Should(HaveKeyWithValue(controllers.AnnoationMembershipSecretCreatedBy, clusterName))
		Expect(secret.Annotations).Should(HaveKeyWithValue(controllers.AnnotationSecretManagedBy, controllers.SecretManagedBy))
		Expect(controllerutil.ContainsFinalizer(secret, controllers.FinalizerMembership))

		data := secret.Data[controllers.SecretKeyGoogleApplicationCredentials]

		var actualMembership gkehubpb.Membership
		Expect(json.Unmarshal(data, &actualMembership)).To(Succeed())

		Expect(actualMembership.Name).To(Equal("/project/the-project/locations/global/membership/the-membership"))
		Expect(actualMembership.Authority.Issuer).To(Equal(membership.KubernetesIssuer))
		Expect(actualMembership.Authority.WorkloadIdentityPool).To(Equal("the-workload-id-pool"))
		Expect(actualMembership.Authority.IdentityProvider).To(Equal("the-identity-provider"))
	})

	When("workload identity is not enabled", func() {
		BeforeEach(func() {
			cluster := gcpCluster.DeepCopy()
			cluster.Annotations = map[string]string{}

			Expect(k8sClient.Update(ctx, cluster)).To(Succeed())
		})

		It("should return an error and skip cluster", func() {
			Expect(reconcilErr).ToNot(HaveOccurred())
		})

		It("should not create a membership secret", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      controllers.MembershipSecretName,
				Namespace: namespace,
			}, secret)

			Expect(err).To(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})

	When("the kubeadm control plane is not ready", func() {
		BeforeEach(func() {
			controlPlaneStatus := capi.KubeadmControlPlaneStatus{
				Ready: false,
			}

			tests.PatchControlPlaneStatus(k8sClient, kubeadmControlPlane, controlPlaneStatus)
		})

		It("requeues the request", func() {
			Expect(reconcilErr).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeTrue())
			Expect(result.RequeueAfter).To(Equal(time.Second * 15))
		})
	})

	When("the workload cluster config is missing", func() {
		BeforeEach(func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-kubeconfig", gcpCluster.Name),
					Namespace: namespace,
				},
			}

			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("returns a not found error", func() {
			Expect(reconcilErr).To(HaveOccurred())
			Expect(k8serrors.IsNotFound(reconcilErr)).To(BeTrue())
		})
	})

	When("the workcload cluster config data is missing", func() {
		BeforeEach(func() {
			secret := kubeconfigSecret.DeepCopy()
			secret.Data = map[string][]byte{
				"some-other-key": []byte("some-data"),
			}

			Expect(k8sClient.Patch(ctx, secret, client.MergeFrom(kubeconfigSecret))).To(Succeed())
		})

		It("returns an error", func() {
			Expect(reconcilErr).To(MatchError(ContainSubstring("cluster kubeconfig data is missing")))
		})
	})

	When("the workload cluster config is broken", func() {
		BeforeEach(func() {
			secret := kubeconfigSecret.DeepCopy()
			secret.Data = map[string][]byte{
				"value": []byte("{'title': 'Its a cold cold world'}"),
			}

			Expect(k8sClient.Patch(ctx, secret, client.MergeFrom(kubeconfigSecret))).To(Succeed())
		})

		It("returns an error", func() {
			Expect(reconcilErr).To(HaveOccurred())
			Expect(clientcmd.IsConfigurationInvalid(reconcilErr)).To(BeTrue())
		})
	})

	When("the membership client fails", func() {
		BeforeEach(func() {
			oops := errors.New("something went wrong")
			fakeGKEClient.RegisterReturns(nil, oops)
		})

		It("should return an error", func() {
			Expect(reconcilErr).To(HaveOccurred())
		})

		It("should not create a membership secret", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      controllers.MembershipSecretName,
				Namespace: namespace,
			}, secret)

			Expect(err).To(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})

	When("the membership secret already exists", func() {
		BeforeEach(func() {
			membershipSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      controllers.MembershipSecretName,
					Namespace: controllers.DefaultMembershipSecretNamespace,
				},
			}

			Expect(k8sClient.Create(ctx, membershipSecret)).To(Succeed())
		})

		It("should not return an error", func() {
			Expect(reconcilErr).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})
	})
})
