package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/hetzner"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type HetznerDNSProviderSolver struct {
	ClientFactory     func(*rest.Config) (kubernetes.Interface, error)
	client            kubernetes.Interface
	DNSClientFactory  func(string, string) hetzner.DNSClient
	DefaultAPIKeyName string
	DefaultAPIKeyKey  string
	BaseURL           string
}

type hetznerDNSProviderConfig struct {
	APIKeySecretRef cmmeta.SecretKeySelector `json:"apiKeySecretRef"`
	ZoneID          string                   `json:"zoneId"`
}

func New(apiKeyName string, apiKeyKey string, baseURL string) HetznerDNSProviderSolver {
	return HetznerDNSProviderSolver{
		ClientFactory:     func(c *rest.Config) (kubernetes.Interface, error) { return kubernetes.NewForConfig(c) },
		DNSClientFactory:  func(s1, s2 string) hetzner.DNSClient { return hetzner.NewDNS(s1, s2) },
		DefaultAPIKeyName: apiKeyName,
		DefaultAPIKeyKey:  apiKeyKey,
		BaseURL:           baseURL,
	}
}

// Name return the name of this webhook.
func (c *HetznerDNSProviderSolver) Name() string {
	return "hetzner"
}

// Initialize will be called when the webhook first starts.
func (c *HetznerDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := c.ClientFactory(kubeClientConfig)
	if err != nil {
		return err
	}
	c.client = cl
	return nil
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
func (c *HetznerDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	ctx := context.TODO()

	cfg, err := c.loadConfig(ch)
	if err != nil {
		return err
	}

	apiKey, err := c.loadAPIKey(ctx, cfg, ch.ResourceNamespace)
	if err != nil {
		return fmt.Errorf("failed to load API key; %v", err)
	}

	recordName, zoneName, err := extractDomain(ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	dns := c.DNSClientFactory(apiKey, c.BaseURL)
	zone := &hetzner.Zone{}
	if cfg.ZoneID != "" {
		zone.ID = cfg.ZoneID
	} else {
		zone, err = dns.LoadZoneByName(ctx, zoneName)
		if err != nil {
			return fmt.Errorf("failed to load zone id; %v", err)
		}
	}
	if zone == nil {
		return fmt.Errorf("failed find zone")
	}

	_, err = dns.CreateRecord(ctx, zone.ID, hetzner.RecordInfo{
		Type:  "TXT",
		Name:  recordName,
		Value: ch.Key,
		TTL:   120,
	})
	if err != nil {
		return fmt.Errorf("failed to create TXT record; %v", err)
	}

	klog.InfoS("presented challenge", "DNSName", ch.DNSName, "UID", ch.UID)
	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
func (c *HetznerDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	ctx := context.TODO()

	cfg, err := c.loadConfig(ch)
	if err != nil {
		return err
	}

	apiKey, err := c.loadAPIKey(ctx, cfg, ch.ResourceNamespace)
	if err != nil {
		return fmt.Errorf("failed to load API key; %v", err)
	}

	dns := c.DNSClientFactory(apiKey, c.BaseURL)

	recordName, zoneName, err := extractDomain(ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	zone, err := dns.LoadZoneByName(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("failed to load zone id; %v", err)
	}
	if zone == nil {
		return fmt.Errorf("failed find zone")
	}

	records, err := dns.LoadRecords(ctx, zone.ID)
	if err != nil {
		return fmt.Errorf("failed to load zone id; %v", err)
	}

	var record hetzner.Record
	for _, r := range records {
		if r.Name == recordName && r.Type == "TXT" {
			record = r
		}
	}

	if record.Name == "" {
		return fmt.Errorf("failed find record")
	}

	err = dns.DeleteRecord(ctx, record.ID)
	if err != nil {
		return fmt.Errorf("failed to delete zone; %v", err)
	}

	klog.InfoS("cleaned up challenge", "DNSName", ch.DNSName, "UID", ch.UID)
	return nil
}

// extractDomain transforms cert managers FQDN into name and domain.
func extractDomain(fqdn string) (string, string, error) {
	r := regexp.MustCompile(`^([^.]+)\.(.+)\.$`)
	m := r.FindStringSubmatch(fqdn)
	if len(m) != 3 {
		return "", "", fmt.Errorf("failed to split FQDN")
	}
	return m[1], m[2], nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func (c *HetznerDNSProviderSolver) loadConfig(ch *v1alpha1.ChallengeRequest) (hetznerDNSProviderConfig, error) {
	// handle the 'base case' where no configuration has been provided
	if ch.Config == nil {
		return hetznerDNSProviderConfig{
			APIKeySecretRef: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference{Name: c.DefaultAPIKeyName},
				Key:                  c.DefaultAPIKeyKey,
			},
		}, nil
	}

	cfg := hetznerDNSProviderConfig{}
	if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal solver config: %v", err)
	}

	if cfg.APIKeySecretRef.Name == "" {
		cfg.APIKeySecretRef.Name = c.DefaultAPIKeyName
	}
	if cfg.APIKeySecretRef.Key == "" {
		cfg.APIKeySecretRef.Key = c.DefaultAPIKeyKey
	}

	return cfg, nil
}

// loadAPIKey loads the DNS api key from a secret
func (c *HetznerDNSProviderSolver) loadAPIKey(ctx context.Context, cfg hetznerDNSProviderConfig, ns string) (string, error) {
	sec, err := c.client.CoreV1().Secrets(ns).Get(ctx, cfg.APIKeySecretRef.Name, metav1.GetOptions{})

	if err != nil {
		return "", fmt.Errorf("unable to get secret `%s/%s`; %v", cfg.APIKeySecretRef.Name, ns, err)
	}

	apiKey, ok := sec.Data[cfg.APIKeySecretRef.Key]
	if !ok {
		return "", fmt.Errorf("key %s missing in secret `%s/%s`", cfg.APIKeySecretRef.Key, cfg.APIKeySecretRef.Name, ns)
	}

	return string(apiKey), nil
}
