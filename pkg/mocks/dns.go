package mocks

import (
	"context"

	"github.com/S-Bohn/cert-manager-webhook-hetzner/pkg/hetzner"
)

type DNSMock struct {
	LoadZoneByNameFunc func(ctx context.Context, name string) (*hetzner.Zone, error)
	CreateRecordFunc   func(ctx context.Context, zoneID string, info hetzner.RecordInfo) (hetzner.Record, error)
	DeleteRecordFunc   func(ctx context.Context, id string) error
	LoadRecordsFunc    func(ctx context.Context, i string) ([]hetzner.Record, error)
}

func (s *DNSMock) CreateRecord(ctx context.Context, zoneID string, info hetzner.RecordInfo) (hetzner.Record, error) {
	return s.CreateRecordFunc(ctx, zoneID, info)
}

func (s *DNSMock) DeleteRecord(ctx context.Context, id string) error {
	return s.DeleteRecordFunc(ctx, id)
}

func (s *DNSMock) LoadZoneByName(ctx context.Context, name string) (*hetzner.Zone, error) {
	return s.LoadZoneByNameFunc(ctx, name)
}

func (s *DNSMock) LoadRecords(ctx context.Context, id string) ([]hetzner.Record, error) {
	return s.LoadRecordsFunc(ctx, id)
}
