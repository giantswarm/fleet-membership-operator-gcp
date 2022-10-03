package membership_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/fleet-membership-operator-gcp/tests"
)

var gcpProject string

func TestMembership(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Membership Suite")
}

var _ = BeforeSuite(func() {
	tests.GetEnvOrSkip("GOOGLE_APPLICATION_CREDENTIALS")
	gcpProject = tests.GetEnvOrSkip("GCP_PROJECT_ID")
})

func generateJWKS() []byte {
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())
	publickey := &raw.PublicKey

	keySet := jwk.NewSet()
	key, err := jwk.FromRaw(publickey)
	Expect(err).NotTo(HaveOccurred())

	Expect(keySet.AddKey(key)).To(Succeed())

	jwks, err := json.Marshal(keySet)
	Expect(err).NotTo(HaveOccurred())

	return jwks
}
