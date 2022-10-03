package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

const (
	AnnotationWorkloadIdentityEnabled  = "giantswarm.io/workload-identity-enabled"
	AnnoationMembershipSecretCreatedBy = "app.kubernetes.io/created-by" //#nosec G101
	AnnotationSecretManagedBy          = "app.kubernetes.io/managed-by" //#nosec  G101

	FinalizerMembership              = "fleet-membership-operator-gcp.giantswarm.io/finalizer"
	SuffixMembershipName             = "workload-identity"
	MembershipSecretName             = "fleet-membership-operator-gcp-membership"
	DefaultMembershipSecretNamespace = "giantswarm"
	KeyWorkloadClusterConfig         = "value"

	SecretManagedBy = "fleet-membership-operator-gcp" //#nosec G101

	SecretKeyGoogleApplicationCredentials = "config"
)

//counterfeiter:generate . GKEMembershipClient
type GKEMembershipClient interface {
	RegisterMembership(ctx context.Context, cluster *capg.GCPCluster, jwks []byte) (*gkehubpb.Membership, error)
}

// GCPClusterReconciler reconciles a GCPCluster object
type GCPClusterReconciler struct {
	client.Client
	Logger logr.Logger

	MembershipSecretNamespace string
	GKEMembershipClient       GKEMembershipClient
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters/finalizers,verbs=update

func (r *GCPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("gcpcluster", req.NamespacedName)

	gcpCluster := &capg.GCPCluster{}

	err := r.Get(ctx, req.NamespacedName, gcpCluster)
	if err != nil {
		logger.Error(err, "could not get gcp cluster")
		return reconcile.Result{}, nil
	}

	if !r.hasWorkloadIdentityEnabled(gcpCluster) {
		message := fmt.Sprintf("skipping Cluster %s because workload identity is not enabled", gcpCluster.Name)
		logger.Info(message)
		return reconcile.Result{}, nil
	}

	if !gcpCluster.Status.Ready {
		message := fmt.Sprintf("skipping Cluster %s because its not yet ready", gcpCluster.Name)
		logger.Info(message)
		return reconcile.Result{}, nil
	}

	kubeadmControlPlane := &capi.KubeadmControlPlane{}
	err = r.Get(ctx, req.NamespacedName, kubeadmControlPlane)
	if err != nil {
		logger.Error(err, "could not get the kubeadm control plane")
		return reconcile.Result{}, err
	}

	if !kubeadmControlPlane.Status.Ready {
		message := fmt.Sprintf("skipping Cluster %s because controlplane is not ready", gcpCluster.Name)
		logger.Info(message)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 15,
		}, nil
	}

	config, err := r.getWorkloadClusterConfig(ctx, gcpCluster, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to get kubeconfig")
		return reconcile.Result{}, err
	}

	workloadClusterClient, err := client.New(config, client.Options{})
	if err != nil {
		logger.Error(err, "failed to create workload cluster client")
		return reconcile.Result{}, err
	}

	logger.Info(fmt.Sprintf("Cluster name is %s", gcpCluster.Name))

	oidcJwks, err := r.getOIDCJWKS(config)
	if err != nil {
		logger.Error(err, "failed to get cluster oidc jwks")
		return reconcile.Result{}, err
	}

	membership, err := r.GKEMembershipClient.RegisterMembership(ctx, gcpCluster, oidcJwks)
	if err != nil {
		logger.Error(err, "failed to reconcile gke membership")
		return reconcile.Result{}, err
	}

	membershipJson, err := json.Marshal(membership)
	if err != nil {
		logger.Error(err, "failed to marshal membership json")
		return reconcile.Result{}, err
	}

	secret := r.generateMembershipSecret(membershipJson, gcpCluster)
	err = workloadClusterClient.Create(ctx, secret)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		logger.Error(err, "failed to create secret on workload cluster")
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GCPClusterReconciler) hasWorkloadIdentityEnabled(cluster *capg.GCPCluster) bool {
	_, exists := cluster.Annotations[AnnotationWorkloadIdentityEnabled]
	return exists
}

func (r *GCPClusterReconciler) getWorkloadClusterConfig(ctx context.Context, cluster *capg.GCPCluster, namespace string) (*rest.Config, error) {
	secret := &corev1.Secret{}
	secretName := fmt.Sprintf("%s-kubeconfig", cluster.Name)

	err := r.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}, secret)
	if err != nil {
		r.Logger.Error(err, "could not get cluster secret")
		return nil, err
	}

	data, ok := secret.Data[KeyWorkloadClusterConfig]
	if !ok {
		err = errors.New("cluster kubeconfig data is missing")
		return nil, err
	}

	config, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return config.ClientConfig()
}

func (r *GCPClusterReconciler) getOIDCJWKS(config *rest.Config) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s/openid/v1/jwks", config.Host)

	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		r.Logger.Error(err, "failed to create http client")
		return []byte{}, err
	}

	resp, err := httpClient.Get(reqUrl)
	if err != nil {
		r.Logger.Error(err, "failed to fetch jwks")
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error(err, "failed to read oidc jwks response body")
		return []byte{}, err
	}

	return body, nil
}

func (r *GCPClusterReconciler) generateMembershipSecret(membershipJson []byte, cluster *capg.GCPCluster) *corev1.Secret {
	membershipJsonString := string(membershipJson)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MembershipSecretName,
			Namespace: r.MembershipSecretNamespace,
			Annotations: map[string]string{
				AnnoationMembershipSecretCreatedBy: cluster.Name,
				AnnotationSecretManagedBy:          SecretManagedBy,
			},
		},
		StringData: map[string]string{
			SecretKeyGoogleApplicationCredentials: membershipJsonString,
		},
	}

	ok := controllerutil.AddFinalizer(secret, FinalizerMembership)
	if !ok {
		message := fmt.Sprintf("failed to add finalizer for %s membership secret", cluster.Name)
		r.Logger.Info(message)
	}

	return secret
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capg.GCPCluster{}).
		Complete(r)
}
