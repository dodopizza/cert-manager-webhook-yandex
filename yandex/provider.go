package yandex

import (
	"context"
	"errors"
	"fmt"
	"strings"

	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/gen/dns"
	"github.com/yandex-cloud/go-sdk/iamkey"

	"github.com/dodopizza/cert-manager-webhook-yandex/yandex/internal"
)

const (
	// AuthorizationTypeInstanceServiceAccount is the authorization type describes that
	// Compute Instance Service Account credentials should be used for authorizing requests to Yandex Cloud
	AuthorizationTypeInstanceServiceAccount = "instance-service-account"

	// AuthorizationTypeOAuthToken is the authorization type describes that
	// OAuth token should be used for authorizing requests to Yandex Cloud
	AuthorizationTypeOAuthToken = "iam-token"

	// AuthorizationTypeKey is the authorization type describes that
	// Service Account authorization key file used for authorizing requests to Yandex Cloud
	AuthorizationTypeKey = "iam-key-file"

	// DNSRecordSetDefaultTTL is the default TTL for record sets
	DNSRecordSetDefaultTTL = int64(60)
)

type DNSProviderConfig struct {
	AuthorizationType       string
	AuthorizationOAuthToken string
	AuthorizationKey        string
	FolderId                string
}

type DNSProvider struct {
	client *dns.DnsZoneServiceClient
	folder string
}

func NewProviderConfigFromValues() *DNSProviderConfig {
	return &DNSProviderConfig{}
}

func NewProviderConfigFromEnv() *DNSProviderConfig {
	return &DNSProviderConfig{}
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

// Validate is a method for checking invariants of DNSProviderConfig
//
// Returns error if any required field is missing, otherwise nil
func (cfg *DNSProviderConfig) Validate() error {
	if cfg.FolderId == "" {
		return errors.New("required field \"FolderId\" is missing")
	}

	authorizationTypes := []string{
		AuthorizationTypeInstanceServiceAccount,
		AuthorizationTypeOAuthToken,
		AuthorizationTypeKey,
	}

	if !internal.ContainsString(cfg.AuthorizationType, authorizationTypes) {
		return errors.New("required field \"AuthorizationType\" is missing")
	}

	if cfg.AuthorizationType == AuthorizationTypeOAuthToken && cfg.AuthorizationOAuthToken == "" {
		return fmt.Errorf("required field \"AuthorizationOAuthToken\" is missing for authorization type: %s",
			cfg.AuthorizationType)
	}

	if cfg.AuthorizationType == AuthorizationTypeKey && cfg.AuthorizationKey == "" {
		return fmt.Errorf("required field \"AuthorizationTypeKey\" is missing for authorization type: %s",
			cfg.AuthorizationType)
	}

	return nil
}

// Helper function
//
// Returns ycsdk.Credentials that resolved by AuthorizationType or an error if operation failed.
func (cfg *DNSProviderConfig) credentials() (ycsdk.Credentials, error) {
	auth := strings.TrimSpace(cfg.AuthorizationType)

	switch auth {
	case AuthorizationTypeInstanceServiceAccount:
		return ycsdk.InstanceServiceAccount(), nil
	case AuthorizationTypeOAuthToken:
		return ycsdk.OAuthToken(cfg.AuthorizationOAuthToken), nil
	case AuthorizationTypeKey:
		key, err := iamkey.ReadFromJSONBytes([]byte(cfg.AuthorizationKey))
		if err != nil {
			return nil, err
		}
		return ycsdk.ServiceAccountKey(key)
	default:
		return nil, errors.New("unsupported authorization type")
	}
}
