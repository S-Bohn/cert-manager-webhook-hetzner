package hetzner_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/internal/hetzner"
	"github.com/matryer/is"
)

type HTTPMockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (s HTTPMockClient) Do(req *http.Request) (*http.Response, error) {
	return s.DoFunc(req)
}

func TestCreateRecord(t *testing.T) {
	is := is.New(t)

	d := hetzner.NewDNS("abc123", "https://dns.hetzner.com/api")
	req := &http.Request{}
	res := &http.Response{}

	d.Client = HTTPMockClient{
		DoFunc: func(r *http.Request) (*http.Response, error) {
			req = r
			return res, nil
		},
	}
	res = &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewReader([]byte(`{"record": {
				"type": "A",
				"id": "the_id",
				"created": "2021-08-18T13:08:19Z",
				"modified": "2021-09-13T10:18:29Z",
				"zone_id": "the_zone_id",
				"name": "a_name",
				"value": "127.0.0.1",
				"ttl": 123
			  }
		  }`))),
	}
	r, e := d.CreateRecord(context.TODO(), "someZoneIdentifier", hetzner.RecordInfo{
		Type:  "AAAA",
		Name:  "mail",
		Value: "127.0.0.1",
		TTL:   1234,
	})
	is.NoErr(e)
	is.Equal(r.ID, "the_id")
	is.Equal(r.Type, "A")
	is.Equal(r.ZoneID, "the_zone_id")
	is.Equal(r.Name, "a_name")
	is.Equal(r.Value, "127.0.0.1")
	is.Equal(req.URL.Host, "dns.hetzner.com")
	is.Equal(req.URL.Path, "/api/v1/records")
	is.Equal(req.Header[http.CanonicalHeaderKey("Auth-Api-Token")][0], "abc123")
}

func TestLoadByName(t *testing.T) {
	is := is.New(t)
	d := hetzner.NewDNS("abc123", "https://dns.hetzner.com/api")
	req := &http.Request{}
	res := &http.Response{}
	err := fmt.Errorf("")
	err = nil

	d.Client = HTTPMockClient{
		DoFunc: func(r *http.Request) (*http.Response, error) {
			req = r
			return res, err
		},
	}
	res = &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(bytes.NewReader([]byte(`
		{
			"zones": [
			  {
				"id": "rXzFZePHbFwsVdRQmKzBdm",
				"name": "example.com",
				"ttl": 86400,
				"registrar": "",
				"legacy_dns_host": "",
				"legacy_ns": [
				  "oxygen.ns.hetzner.com.",
				  "helium.ns.hetzner.de.",
				  "hydrogen.ns.hetzner.com."
				],
				"ns": [
				  "hydrogen.ns.hetzner.com",
				  "oxygen.ns.hetzner.com",
				  "helium.ns.hetzner.de"
				],
				"created": "2021-02-28 17:11:18.007 +0000 UTC",
				"verified": "",
				"modified": "2021-02-28 17:11:18.72 +0000 UTC",
				"project": "proj",
				"owner": "own",
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
				"records_count": 4
			  }
			],
			"meta": {
			  "pagination": {
				"page": 1,
				"per_page": 100,
				"previous_page": 1,
				"next_page": 1,
				"last_page": 1,
				"total_entries": 3
			  }
			}
		  }`))),
	}

	z, e := d.LoadZoneByName(context.TODO(), "my_zone_name")
	is.NoErr(e)
	is.Equal(z.ID, "rXzFZePHbFwsVdRQmKzBdm")
	is.Equal(z.Name, "example.com")
	is.Equal(req.URL.Query()["name"][0], "my_zone_name")
	is.Equal(req.Header[http.CanonicalHeaderKey("Auth-Api-Token")][0], "abc123")
}

func TestDeleteRecord(t *testing.T) {
	is := is.New(t)
	d := hetzner.NewDNS("abc123", "https://dns.hetzner.com/api")
	req := &http.Request{}
	res := &http.Response{}
	err := fmt.Errorf("")
	err = nil

	d.Client = HTTPMockClient{
		DoFunc: func(r *http.Request) (*http.Response, error) {
			req = r
			return res, err
		},
	}
	res = &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(``))),
	}

	e := d.DeleteRecord(context.TODO(), "Z0n31dz0Ne")
	is.NoErr(e)
	is.Equal(req.URL.Path, "/api/v1/records/Z0n31dz0Ne")
	is.Equal(req.Header[http.CanonicalHeaderKey("Auth-Api-Token")][0], "abc123")
}
