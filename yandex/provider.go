package yandex

import (
	"context"
	"fmt"

	dnsProto "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/gen/dns"
)

type DNSProvider struct {
	client *dns.DnsZoneServiceClient
	config *DNSProviderConfig
}

func NewProvider(cfg *DNSProviderConfig) (*DNSProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	creds, err := cfg.credentials()
	if err != nil {
		return nil, err
	}

	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}

	provider := &DNSProvider{
		config: cfg,
		client: sdk.DNS().DnsZone(),
	}
	return provider, nil
}

func (p *DNSProvider) Present(zone, fqdn, key string) error {
	zoneState, err := p.zone(zone)
	if err != nil {
		return err
	}

	return p.addRecord(zoneState.Id, fqdn, key)
}

func (p *DNSProvider) CleanUp(zone, fqdn, key string) error {
	zoneState, err := p.zone(zone)
	if err != nil {
		return err
	}

	return p.removeRecord(zoneState.Id, fqdn, key)
}

func (p *DNSProvider) zone(zone string) (*dnsProto.DnsZone, error) {
	iterator := p.client.DnsZoneIterator(
		context.Background(),
		&dnsProto.ListDnsZonesRequest{
			FolderId: p.config.FolderId,
		},
	)

	for iterator.Next() {
		value := iterator.Value()

		if value.Zone == zone {
			return value, nil
		}
	}

	if iterator.Error() != nil {
		return nil, iterator.Error()
	}

	return nil, fmt.Errorf("zone %s not exists in folder %s", zone, p.config.FolderId)
}

func (p *DNSProvider) addRecord(zoneId, fqdn, key string) error {
	_, err := p.client.UpsertRecordSets(
		context.Background(),
		&dnsProto.UpsertRecordSetsRequest{
			DnsZoneId: zoneId,
			Merges: []*dnsProto.RecordSet{
				{
					Name: fqdn,
					Type: "TXT",
					Ttl:  int64(p.config.DNSRecordSetTTL),
					Data: []string{key},
				},
			},
		},
	)
	return err
}

func (p *DNSProvider) removeRecord(zoneId, fqdn, key string) error {
	_, err := p.client.UpsertRecordSets(
		context.Background(),
		&dnsProto.UpsertRecordSetsRequest{
			DnsZoneId: zoneId,
			Deletions: []*dnsProto.RecordSet{
				{
					Name: fqdn,
					Type: "TXT",
					Ttl:  int64(p.config.DNSRecordSetTTL),
					Data: []string{key},
				},
			},
		},
	)
	return err
}
