package main

import (
	"flag"
	"os"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/internal/webhook"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
)

var (
	groupName         string = os.Getenv("GROUP_NAME")
	apiBaseURL        string = os.Getenv("DNS_API_URL")
	defaultAPIKeyKey  string = os.Getenv("DNS_API_DEFAULT_SECRET_KEY")
	defaultAPIKeyName string = os.Getenv("DNS_API_DEFAULT_SECRET_NAME")
)

func main() {
	flag.StringVar(&groupName, "group-name", groupName, "define the group name")
	flag.StringVar(&apiBaseURL, "api-base-url", apiBaseURL, "override hetzner dns api base url")
	flag.StringVar(&defaultAPIKeyName, "api-key-secret-name", defaultAPIKeyName, "allows setting a default secret for the hetzner dns api key")
	flag.StringVar(&defaultAPIKeyKey, "api-key-secret-key", defaultAPIKeyKey, "allows setting a default secret key for the hetzner dns api key")

	flag.Parse()

	if groupName == "" {
		panic("GROUP_NAME must be specified")
	}

	if apiBaseURL == "" {
		apiBaseURL = "https://dns.hetzner.com/api"
	}

	w := webhook.New(defaultAPIKeyName, defaultAPIKeyKey, apiBaseURL)
	cmd.RunWebhookServer(groupName, &w)
}
