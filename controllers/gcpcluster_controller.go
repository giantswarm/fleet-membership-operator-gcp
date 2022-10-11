package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/fleet-membership-operator-gcp/pkg/workload"
	"github.com/giantswarm/fleet-membership-operator-gcp/types"

	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

const (
	AnnotationWorkloadIdentityEnabled = "giantswarm.io/workload-identity-enabled"

	FinalizerMembership  = "fleet-membership-operator-gcp.giantswarm.io/finalizer"
	SuffixMembershipName = "workload-identity"
)

//counterfeiter:generate . GKEMembershipClient
type GKEMembershipClient interface {
	Register(ctx context.Context, cluster *capg.GCPCluster, jwks []byte) (types.MembershipData, error)
	Unregister(ctx context.Context, cluster *capg.GCPCluster) error
}

// GCPClusterReconciler reconciles a GCPCluster object
type GCPClusterReconciler struct {
	membershipSecretNamespace string

	runtimeClient       client.Client
	gkeMembershipClient GKEMembershipClient
	clusterConfigs      *workload.ClusterConfigs
}

func NewGCPClusterReconciler(membershipSecretNamespace string, runtimeClient client.Client, membershipClient GKEMembershipClient) *GCPClusterReconciler {
	return &GCPClusterReconciler{
		runtimeClient:             runtimeClient,
		membershipSecretNamespace: membershipSecretNamespace,
		gkeMembershipClient:       membershipClient,
		clusterConfigs:            workload.NewClusterConfigs(runtimeClient),
	}
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpclusters/finalizers,verbs=update

func (r *GCPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger = logger.WithName("gcpcluster-reconciler")
	logger.Info("Reconciling cluster")
	defer logger.Info("Finished reconciling cluster")

	gcpCluster := &capg.GCPCluster{}
	err := r.runtimeClient.Get(ctx, req.NamespacedName, gcpCluster)
	if err != nil {
		logger.Error(err, "could not get gcp cluster")
		return reconcile.Result{}, nil
	}

	if !r.hasWorkloadIdentityEnabled(gcpCluster) {
		message := fmt.Sprintf("skipping Cluster %s because workload identity is not enabled", gcpCluster.Name)
		logger.Info(message)
		return reconcile.Result{}, nil
	}

	if !gcpCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, gcpCluster)
	}

	return r.reconcileNormal(ctx, logger, gcpCluster)
}

func (r *GCPClusterReconciler) reconcileDelete(ctx context.Context, logger logr.Logger, gcpCluster *capg.GCPCluster) (reconcile.Result, error) {
	err := r.gkeMembershipClient.Unregister(ctx, gcpCluster)
	if err != nil {
		logger.Error(err, "failed to unregister cluster membership")
		return reconcile.Result{}, err
	}

	err = r.removeFinalizer(ctx, gcpCluster)
	if err != nil {
		logger.Error(err, "failed to add finalizer to cluster")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *GCPClusterReconciler) reconcileNormal(ctx context.Context, logger logr.Logger, cluster *capg.GCPCluster) (reconcile.Result, error) {
	if !cluster.Status.Ready {
		message := fmt.Sprintf("skipping Cluster %s because its not yet ready", cluster.Name)
		logger.Info(message)
		return reconcile.Result{}, nil
	}

	kubeadmControlPlane := &capi.KubeadmControlPlane{}
	nsName := k8stypes.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}
	err := r.runtimeClient.Get(ctx, nsName, kubeadmControlPlane)
	if err != nil {
		logger.Error(err, "could not get the kubeadm control plane")
		return reconcile.Result{}, err
	}

	if !kubeadmControlPlane.Status.Ready {
		message := fmt.Sprintf("skipping Cluster %s because controlplane is not ready", cluster.Name)
		logger.Info(message)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 15,
		}, nil
	}

	return r.registerMembership(ctx, logger, cluster)
}

func (r *GCPClusterReconciler) registerMembership(ctx context.Context, logger logr.Logger, cluster *capg.GCPCluster) (ctrl.Result, error) {
	config, err := r.clusterConfigs.Get(ctx, cluster)
	if err != nil {
		logger.Error(err, "failed to get cluster kubeconfig")
		return reconcile.Result{}, err
	}

	workloadCluster, err := workload.NewCluster(config)
	if err != nil {
		logger.Error(err, "failed to create workload cluster client")
		return reconcile.Result{}, err
	}

	oidcJwks, err := workloadCluster.GetOIDCJWKS()
	if err != nil {
		logger.Error(err, "failed to get cluster oidc jwks")
		return reconcile.Result{}, err
	}

	err = r.addFinalizer(ctx, cluster)
	if err != nil {
		logger.Error(err, "failed to add finalizer to cluster")
		return reconcile.Result{}, err
	}

	membership, err := r.gkeMembershipClient.Register(ctx, cluster, oidcJwks)
	if err != nil {
		logger.Error(err, "failed to reconcile gke membership")
		return reconcile.Result{}, err
	}

	err = workloadCluster.SaveMembershipData(ctx, r.membershipSecretNamespace, membership)
	if err != nil {
		logger.Error(err, "failed to save membership data in workload cluster")
		return reconcile.Result{}, err

	}

	return ctrl.Result{}, nil
}

func (r *GCPClusterReconciler) addFinalizer(ctx context.Context, cluster *capg.GCPCluster) error {
	originalCluster := cluster.DeepCopy()
	controllerutil.AddFinalizer(cluster, FinalizerMembership)
	return r.runtimeClient.Patch(ctx, cluster, client.MergeFrom(originalCluster))
}

func (r *GCPClusterReconciler) removeFinalizer(ctx context.Context, cluster *capg.GCPCluster) error {
	originalCluster := cluster.DeepCopy()
	controllerutil.RemoveFinalizer(cluster, FinalizerMembership)
	return r.runtimeClient.Patch(ctx, cluster, client.MergeFrom(originalCluster))
}

func (r *GCPClusterReconciler) hasWorkloadIdentityEnabled(cluster *capg.GCPCluster) bool {
	_, exists := cluster.Annotations[AnnotationWorkloadIdentityEnabled]
	return exists
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capg.GCPCluster{}).
		Complete(r)
}
