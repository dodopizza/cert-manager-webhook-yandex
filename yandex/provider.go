package yandex

import (
	"context"

	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/gen/dns"
)

type DNSProvider struct {
	client *dns.DnsZoneServiceClient
	folder string
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
		folder: cfg.FolderId,
		client: sdk.DNS().DnsZone(),
	}
	return provider, nil
}

func (*DNSProvider) Present(zone, fqdn, key string) error {
	return nil
}

func (*DNSProvider) CleanUp(zone, fqdn, key string) error {
	return nil
}
