package webhook_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/hetzner"
	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/mocks"
	hw "github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/webhook"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestName(t *testing.T) {
	is := is.New(t)
	w := hw.New("key-name", "key-key", "http://dns.hetzner.de/api")

	is.Equal(w.Name(), "hetzner") // Name must return correct data
}

func TestInitialize(t *testing.T) {
	is := is.New(t)
	w := hw.New("key-name", "key-key", "http://dns.hetzner.de/api")
	factoryCalled := false

	w.ClientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		factoryCalled = true
		return &kubernetes.Clientset{}, nil
	}
	err := w.Initialize(&rest.Config{}, make(<-chan struct{}))

	is.True(factoryCalled) // factory must be called
	is.NoErr(err)          // Initialize must not fail

}

func TestInitializeFail(t *testing.T) {
	is := is.New(t)
	w := hw.New("key-name", "key-key", "http://dns.hetzner.de/api")
	factoryCalled := false

	w.ClientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		factoryCalled = true
		return nil, fmt.Errorf("failed")
	}

	err := w.Initialize(&rest.Config{}, make(<-chan struct{}))

	is.True(factoryCalled) // factore must be called
	is.True(err != nil)    // Initialize should fail
}

func TestPresent(t *testing.T) {
	is := is.New(t)
	isr := is.NewRelaxed(t)

	w := hw.New("key-name", "key-key", "https://localhost/api")
	factoryCalled := false
	w.ClientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		factoryCalled = true
		return testclient.NewSimpleClientset(&v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Namespace: meta_v1.NamespaceDefault,
				Name:      "key-name",
			},
			Data: map[string][]byte{
				"key-key": []byte("some-api-key"),
			}}), nil
	}
	w.DNSClientFactory = func(s1, s2 string) hetzner.DNSClient {
		return &mocks.DNSMock{
			LoadZoneByNameFunc: func(ctx context.Context, name string) (*hetzner.Zone, error) {
				isr.Equal(name, "example.org") // id of loaded zone must match

				return &hetzner.Zone{
					Name: name,
					ID:   "xxxZZZ111",
				}, nil
			},
			CreateRecordFunc: func(ctx context.Context, zoneID string, info hetzner.RecordInfo) (hetzner.Record, error) {
				is.Equal(zoneID, "xxxZZZ111")             // zone id of created record must match
				is.Equal(info.Type, "TXT")                // type of created record must match
				is.Equal(info.Name, "_acme-challenge")    // name of created record must match
				is.Equal(info.Value, "ABCsecretlySigned") // value of created record must match

				return hetzner.Record{
					Type:   info.Type,
					ID:     "R3c0RdiD",
					ZoneID: zoneID,
					Name:   info.Name,
					Value:  info.Value,
				}, nil
			},
		}
	}
	err := w.Initialize(&rest.Config{}, make(<-chan struct{}))

	isr.True(factoryCalled) // factory must be called
	is.NoErr(err)           // Initialize must not fail

	err = w.Present(&v1alpha1.ChallengeRequest{
		ResourceNamespace:       "default",
		UID:                     "keeMie5akeeMie5a",
		Action:                  "present",
		Type:                    "dns01",
		DNSName:                 "hetzner.example.org",
		Key:                     "ABCsecretlySigned",
		ResolvedFQDN:            "_acme-challenge.example.org.",
		ResolvedZone:            "example.org.",
		AllowAmbientCredentials: false,
		Config:                  &v1beta1.JSON{Raw: []byte("{}")},
	})
	is.NoErr(err) // Present must not fail
}

func TestCleanUp(t *testing.T) {
	is := is.New(t)
	isr := is.NewRelaxed(t)

	w := hw.New("key-name", "key-key", "https://localhost/api")
	factoryCalled := false
	w.ClientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		factoryCalled = true
		return testclient.NewSimpleClientset(&v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Namespace: meta_v1.NamespaceDefault,
				Name:      "key-name",
			},
			Data: map[string][]byte{
				"key-key": []byte("some-api-key"),
			}}), nil
	}
	w.DNSClientFactory = func(s1, s2 string) hetzner.DNSClient {
		return &mocks.DNSMock{
			LoadZoneByNameFunc: func(ctx context.Context, name string) (*hetzner.Zone, error) {
				isr.Equal(name, "example.org") // loaded zone name must match

				return &hetzner.Zone{
					Name: name,
					ID:   "xxxZZZ111",
				}, nil
			},
			DeleteRecordFunc: func(ctx context.Context, id string) error {
				is.Equal(id, "R3c0RdiD") // deleted record id must match
				return nil
			},
			LoadRecordsFunc: func(ctx context.Context, id string) ([]hetzner.Record, error) {
				is.Equal(id, "xxxZZZ111") // queried zone id must match
				return []hetzner.Record{
					{
						Type:   "A",
						ID:     "WR0N51D",
						ZoneID: "xxxZZZ111",
						Name:   "some_host",
						Value:  "8.8.8.8",
					},
					{
						Type:   "TXT",
						ID:     "R3c0RdiD",
						ZoneID: "xxxZZZ111",
						Name:   "_acme-challenge",
						Value:  "ABCsecretlySigned",
					},
				}, nil
			},
		}
	}
	err := w.Initialize(&rest.Config{}, make(<-chan struct{}))
	isr.True(factoryCalled) // factory must be called
	is.NoErr(err)           // Initialize must not fail

	err = w.CleanUp(&v1alpha1.ChallengeRequest{
		ResourceNamespace:       "default",
		UID:                     "keeMie5akeeMie5a",
		Action:                  "present",
		Type:                    "dns01",
		DNSName:                 "example.org",
		Key:                     "ABCsecretlySigned",
		ResolvedFQDN:            "_acme-challenge.example.org.",
		ResolvedZone:            "example.org.",
		AllowAmbientCredentials: false,
		Config:                  &v1beta1.JSON{Raw: []byte("{}")},
	})
	is.NoErr(err) // CleanUp must not fail
}
