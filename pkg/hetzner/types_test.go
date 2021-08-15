package hetzner_test

import (
	"encoding/json"
	"testing"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/hetzner"
	"github.com/matryer/is"
)

func TestUnmarschalZone(t *testing.T) {
	is := is.New(t)
	res := hetzner.Zone{}
	err := json.Unmarshal([]byte(`
	{
		"id": "AKjLJeqtxEgVGNhJRkGum8",
		"name": "foobar.io",
		"ttl": 86400,
		"registrar": "",
		"legacy_dns_host": "",
		"legacy_ns": [
			"robotns3.second-ns.com.",
			"ns1.first-ns.de.",
			"robotns2.second-ns.de."
		],
		"ns": [
			"hydrogen.ns.hetzner.com",
			"oxygen.ns.hetzner.com",
			"helium.ns.hetzner.de"
		],
		"created": "2020-04-07 01:23:52 +0000 UTC",
		"verified": "2020-04-07 01:54:52.607151685 +0000 UTC m=+684.733523994",
		"modified": "2020-09-30 20:44:25.862 +0000 UTC",
		"project": "",
		"owner": "",
		"permission": "",
		"zone_type": {
			"id": "",
			"name": "",
			"description": "",
			"prices": null
		},
		"status": "verified",
		"paused": false,
		"is_secondary_dns": false,
		"txt_verification": {
			"name": "",
			"token": ""
		},
		"records_count": 51
	}`), &res)
	is.NoErr(err)
	is.Equal(res.ID, "AKjLJeqtxEgVGNhJRkGum8")
	is.Equal(res.Name, "foobar.io")
}
