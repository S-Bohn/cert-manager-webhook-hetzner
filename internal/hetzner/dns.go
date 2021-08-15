package hetzner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"k8s.io/klog/v2"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DNSClient interface {
	LoadZoneByName(ctx context.Context, name string) (*Zone, error)
	CreateRecord(ctx context.Context, zoneID string, info RecordInfo) (Record, error)
	DeleteRecord(ctx context.Context, id string) error
	LoadRecords(ctx context.Context, id string) ([]Record, error)
}

type DNS struct {
	Client      HTTPClient
	ApiKey      string
	ApiEndpoint string
}

// NewDNS creates a DNS struct from given key and endpoint.
func NewDNS(key string, endpoint string) *DNS {
	return &DNS{
		ApiKey:      key,
		Client:      &http.Client{},
		ApiEndpoint: endpoint,
	}
}

// LoadZoneByName loads a DNS zone by given name.
func (s *DNS) LoadZoneByName(ctx context.Context, name string) (*Zone, error) {
	d, err := s.hetznerCall(ctx, "GET", fmt.Sprintf("%s/v1/zones?name=%s", s.ApiEndpoint, url.QueryEscape(name)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNS zones; %v", err)
	}

	res := getAllZonesResponse{}
	err = json.Unmarshal(d, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal getAllZonesResponse; %v", err)
	}

	if len(res.Zones) == 1 {
		return &res.Zones[0], nil
	} else {
		return nil, nil
	}
}

func (s *DNS) CreateRecord(ctx context.Context, zoneID string, info RecordInfo) (Record, error) {
	jsonData, err := json.Marshal(createRecordRequest{
		ZoneID: zoneID,
		Type:   info.Type,
		Name:   info.Name,
		Value:  info.Value,
		TTL:    info.TTL,
	})
	if err != nil {
		return Record{}, fmt.Errorf("failed to marshal createRecord; %v", err)
	}

	resData, err := s.hetznerCall(ctx, "POST", fmt.Sprintf("%s/v1/records", s.ApiEndpoint), bytes.NewBuffer(jsonData))
	if err != nil {
		return Record{}, fmt.Errorf("failed to create DNS record; %v", err)
	}

	res := createRecordResponse{}
	err = json.Unmarshal(resData, &res)
	if err != nil {
		return Record{}, fmt.Errorf("failed to unmarshal createRecordResponse; %v", err)
	}

	return res.Record, nil
}

func (s *DNS) DeleteRecord(ctx context.Context, id string) error {
	_, err := s.hetznerCall(ctx, "DELETE", fmt.Sprintf("%s/v1/records/%s", s.ApiEndpoint, url.QueryEscape(id)), nil)
	if err != nil {
		return fmt.Errorf("failed to delete record %s; %v", id, err)
	}
	return nil
}

func (s *DNS) LoadRecords(ctx context.Context, id string) ([]Record, error) {
	raw, err := s.hetznerCall(ctx, "GET", fmt.Sprintf("%s/v1/records?zone_id=%s", s.ApiEndpoint, url.QueryEscape(id)), nil)
	if err != nil {
		return []Record{}, fmt.Errorf("failed to fetch records from zone %s; %v", id, err)
	}

	res := getAllRecordsResponse{}
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return []Record{}, fmt.Errorf("failed to unmarshal getAllRecordsResponse; %v", err)
	}

	return res.Records, nil
}

// hetznerCall sends an authenticated HTTP request to the given URL.
func (s *DNS) hetznerCall(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return []byte{}, fmt.Errorf("failed initializing request for url %s, method %s; %v", url, method, err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-API-Token", s.ApiKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed; %v", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			klog.ErrorS(err, "failed to close reader", "URL", url, "Method", method)
			os.Exit(255)
		}
	}()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return respBody, nil
	}

	reqBody, _ := io.ReadAll(body)
	err = fmt.Errorf("HTTP %s request to %s failed with status %s", method, url, resp.Status)
	klog.ErrorS(err, "HTTP request failed", "Status", resp.Status,
		"URL", url, "Method", method, "Body", respBody, "Header", resp.Header, "reqBody", reqBody)
	return nil, err
}
