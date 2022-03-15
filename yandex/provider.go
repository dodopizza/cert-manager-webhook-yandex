package yandex

type DNSProvider struct {
}

type DNSProviderConfig struct {
}

func NewProvider(config *DNSProviderConfig) *DNSProvider {
	return &DNSProvider{}
}

func (*DNSProvider) Present(zone, fqdn, key string) error {
	return nil
}

func (*DNSProvider) CleanUp(zone, fqdn, key string) error {
	return nil
}
