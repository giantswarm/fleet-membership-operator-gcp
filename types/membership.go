package types

// The MembershipData is the json representation of the membership, saved on
// the workload cluster
type MembershipData struct {
	Name                 string
	IdentityProvider     string `json:"identity_provider"`
	WorkloadIdentityPool string `json:"workload_identity_pool"`
}
