package yandex

import (
	"errors"
	"fmt"
	"os"

	ycsdk "github.com/yandex-cloud/go-sdk"
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
	AuthorizationTypeKey = "iam-key"
)

const (
	// EnvironmentNamespace is a shared prefix for all environment configuration values
	EnvironmentNamespace = "YANDEX_"

	EnvironmentAuthorizationType       = EnvironmentNamespace + "AUTHORIZATION_TYPE"
	EnvironmentAuthorizationOAuthToken = EnvironmentNamespace + "AUTHORIZATION_OAUTH_TOKEN"
	EnvironmentAuthorizationKey        = EnvironmentNamespace + "AUTHORIZATION_KEY"
	EnvironmentFolderId                = EnvironmentNamespace + "FOLDER_ID"
	EnvironmentDNSRecordSetTTL         = EnvironmentNamespace + "DNS_RECORDSET_TTL"
)

const (
	// DefaultDNSRecordSetTTL is the default TTL for record sets
	DefaultDNSRecordSetTTL = 60

	// DefaultAuthorizationType is the default authorization type for dns api
	DefaultAuthorizationType = AuthorizationTypeInstanceServiceAccount
)

// DNSProviderConfig is used to configure the creation of the DNSProvider.
type DNSProviderConfig struct {
	AuthorizationType       string
	AuthorizationOAuthToken string
	AuthorizationKey        string
	FolderId                string
	DNSRecordSetTTL         int
}

// NewProviderConfig returns a new instance of DNSProviderConfig with specified AuthorizationType and FolderId.
func NewProviderConfig(authorizationType, folderId string) *DNSProviderConfig {
	return &DNSProviderConfig{
		AuthorizationType: authorizationType,
		FolderId:          folderId,
		DNSRecordSetTTL:   DefaultDNSRecordSetTTL,
	}
}

// NewProviderConfigFromEnv returns a new instance of DNSProviderConfig configured from environment variables.
func NewProviderConfigFromEnv() *DNSProviderConfig {
	return &DNSProviderConfig{
		AuthorizationType:       internal.GetEnvOrDefaultString(EnvironmentAuthorizationType, DefaultAuthorizationType),
		AuthorizationOAuthToken: os.Getenv(EnvironmentAuthorizationOAuthToken),
		AuthorizationKey:        os.Getenv(EnvironmentAuthorizationKey),
		FolderId:                os.Getenv(EnvironmentFolderId),
		DNSRecordSetTTL:         internal.GetEnvOrDefaultInt(EnvironmentDNSRecordSetTTL, DefaultDNSRecordSetTTL),
	}
}

// SetSecret sets corresponding secret for DNSProvider depending on AuthorizationType.
func (cfg *DNSProviderConfig) SetSecret(secret string) {
	switch cfg.AuthorizationType {
	case AuthorizationTypeOAuthToken:
		cfg.AuthorizationOAuthToken = secret
	case AuthorizationTypeKey:
		cfg.AuthorizationKey = secret
	}
}

// Validate checks invariants of DNSProviderConfig
// if any required field is missing or incorrect will return error, otherwise nil.
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

	if cfg.DNSRecordSetTTL < DefaultDNSRecordSetTTL {
		return fmt.Errorf("field \"DNSRecordSetTTL\" must be greater or equal to %d", DefaultDNSRecordSetTTL)
	}

	return nil
}

// Helper function
//
// Returns ycsdk.Credentials that resolved by AuthorizationType or an error if operation failed.
func (cfg *DNSProviderConfig) credentials() (ycsdk.Credentials, error) {
	switch cfg.AuthorizationType {
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
