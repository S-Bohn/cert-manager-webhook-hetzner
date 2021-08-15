package main

import (
	"flag"
	"os"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/webhook"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
)

var (
	defaultAPIKeyKey  string
	defaultAPIKeyName string
	groupName         string
	baseURL           string
)

func main() {
	flag.StringVar(&defaultAPIKeyName, "hetzner-default-api-secret-name", "hetzner-secret", "Use this Secret name as default when loading the API key from a Kubernetes Secret.")
	flag.StringVar(&defaultAPIKeyKey, "hetzner-default-api-secret-key", "api-key", "Use this key as default when loading the API key from a Kubernetes Secret.")
	flag.StringVar(&groupName, "group-name", os.Getenv("GROUP_NAME"), "Define the GroupName.")
	flag.StringVar(&baseURL, "hetzner-dns-api-url", "https://dns.hetzner.com/api", "Set the URL to contact Hetzner DNS API.")
	flag.Parse()

	s := webhook.New(defaultAPIKeyName, defaultAPIKeyKey, baseURL)
	cmd.RunWebhookServer(groupName, &s)
}
