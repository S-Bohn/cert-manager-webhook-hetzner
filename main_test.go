package main

import (
	"os"
	"testing"

	hwebhook "github.com/S-Bohn/cert-manager-webhook-hetzner/internal/webhook"
	"github.com/jetstack/cert-manager/test/acme/dns"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.
	//

	// Uncomment the below fixture when implementing your custom DNS provider
	h := hwebhook.New("hetzner-secret", "api-key", "https://dns.hetzner.com/api")
	fixture := dns.NewFixture(&h,
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/hetzner"),
		dns.SetBinariesPath("_test/kubebuilder/bin"),
	)

	/*
		solver := example.New("59351")
		fixture := dns.NewFixture(solver,
			dns.SetResolvedZone("example.com."),
			dns.SetManifestPath("testdata/my-custom-solver"),
			dns.SetBinariesPath("_test/kubebuilder/bin"),
			dns.SetDNSServer("127.0.0.1:59351"),
			dns.SetUseAuthoritative(false),
		)
	*/

	fixture.RunConformance(t)
}
