package yandex

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func unset(envs map[string]string) {
	for env := range envs {
		_ = os.Unsetenv(env)
	}
}

func set(envs map[string]string) {
	for key, value := range envs {
		_ = os.Setenv(key, value)
	}
}

func TestNewProviderConfigFromEnv_DefaultAuth(t *testing.T) {
	type expectedConfig struct {
		folderId           string
		authorizationType  string
		authorizationKey   string
		authorizationToken string
		dnsRecordSetTTL    int
	}

	testCases := []struct {
		name     string
		envs     map[string]string
		expected expectedConfig
	}{
		{
			name: "default options",
			envs: map[string]string{
				EnvironmentFolderId: "bXXXXXXXXXXXXXXXXXXX",
			},
			expected: expectedConfig{
				folderId:           "bXXXXXXXXXXXXXXXXXXX",
				authorizationType:  DefaultAuthorizationType,
				authorizationKey:   "",
				authorizationToken: "",
				dnsRecordSetTTL:    DefaultDNSRecordSetTTL,
			},
		},
		{
			name: "authorization key and ttl specified",
			envs: map[string]string{
				EnvironmentFolderId:          "bXXXXXXXXXXXXXXXXXXX",
				EnvironmentAuthorizationType: AuthorizationTypeKey,
				EnvironmentAuthorizationKey:  "authKey",
				EnvironmentDNSRecordSetTTL:   "123",
			},
			expected: expectedConfig{
				folderId:           "bXXXXXXXXXXXXXXXXXXX",
				authorizationType:  AuthorizationTypeKey,
				authorizationKey:   "authKey",
				authorizationToken: "",
				dnsRecordSetTTL:    123,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set(tc.envs)
			defer unset(tc.envs)

			config := NewProviderConfigFromEnv()

			assert.Equal(t, tc.expected.authorizationType, config.AuthorizationType)
			assert.Equal(t, tc.expected.authorizationToken, config.AuthorizationOAuthToken)
			assert.Equal(t, tc.expected.authorizationKey, config.AuthorizationKey)
			assert.Equal(t, tc.expected.folderId, config.FolderId)
			assert.Equal(t, tc.expected.dnsRecordSetTTL, config.DNSRecordSetTTL)
		})
	}
}

func TestSetSecret(t *testing.T) {
	var config *DNSProviderConfig

	config = NewProviderConfig(AuthorizationTypeKey, "bXXXXXXXXXXXXXXXXXXX")
	config.SetSecret("1234")
	assert.Equal(t, "1234", config.AuthorizationKey)

	config = NewProviderConfig(AuthorizationTypeOAuthToken, "bXXXXXXXXXXXXXXXXXXX")
	config.SetSecret("1234")
	assert.Equal(t, "1234", config.AuthorizationOAuthToken)
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name                 string
		envs                 map[string]string
		expectedErrorMessage string
	}{
		{
			name:                 "empty folder id",
			envs:                 map[string]string{},
			expectedErrorMessage: "\"FolderId\"",
		},
		{
			name: "invalid authorization type",
			envs: map[string]string{
				EnvironmentFolderId:          "bXXXXXXXXXXXXXXXXXXX",
				EnvironmentAuthorizationType: "some",
			},
			expectedErrorMessage: "\"AuthorizationType\"",
		},
		{
			name: "authorization type token with token missing",
			envs: map[string]string{
				EnvironmentFolderId:          "bXXXXXXXXXXXXXXXXXXX",
				EnvironmentAuthorizationType: AuthorizationTypeOAuthToken,
			},
			expectedErrorMessage: "\"AuthorizationOAuthToken\"",
		},
		{
			name: "authorization type key with key missing",
			envs: map[string]string{
				EnvironmentFolderId:          "bXXXXXXXXXXXXXXXXXXX",
				EnvironmentAuthorizationType: AuthorizationTypeKey,
			},
			expectedErrorMessage: "\"AuthorizationTypeKey\"",
		},
		{
			name: "ttl record less than minimum",
			envs: map[string]string{
				EnvironmentFolderId:          "bXXXXXXXXXXXXXXXXXXX",
				EnvironmentAuthorizationType: AuthorizationTypeKey,
				EnvironmentAuthorizationKey:  "key",
				EnvironmentDNSRecordSetTTL:   "23",
			},
			expectedErrorMessage: "\"DNSRecordSetTTL\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set(tc.envs)
			defer unset(tc.envs)

			err := NewProviderConfigFromEnv().Validate()

			assert.Matches(t, err.Error(), tc.expectedErrorMessage)
		})
	}
}
