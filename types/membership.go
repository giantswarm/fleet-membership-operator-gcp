package types

type Membership struct {
	IdentityProvider     string `json:"identity_provider"`
	WorkloadIdentityPool string `json:"workload_identity_pool"`
}
