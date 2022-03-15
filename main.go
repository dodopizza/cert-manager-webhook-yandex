package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	core "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"

	"github.com/dodopizza/cert-manager-webhook-yandex/yandex"
)

var (
	GroupName = os.Getenv("GROUP_NAME")
)

const (
	ProviderName = "yandex"
)

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName, &yandexDNSProviderSolver{})
}

// yandexDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your Yandex Cloud DNS provider.
// This is the implementation of `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type yandexDNSProviderSolver struct {
	client *kubernetes.Clientset
}

// yandexDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
type yandexDNSProviderConfig struct {
	APIKeySecretRef   core.SecretKeySelector `json:"apiKeySecretRef"`
	AuthorizationType string                 `json:"authorizationType"`
	FolderId          string                 `json:"folderId"`
	DNSRecordSetTTL   int                    `json:"dnsRecordSetTTL"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
func (*yandexDNSProviderSolver) Name() string {
	return ProviderName
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (s *yandexDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	provider, err := s.provider(cfg, ch)
	if err != nil {
		return err
	}

	return provider.Present(ch.ResolvedZone, ch.ResolvedFQDN, ch.Key)
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (s *yandexDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	provider, err := s.provider(cfg, ch)
	if err != nil {
		return err
	}

	return provider.CleanUp(ch.ResolvedZone, ch.ResolvedFQDN, ch.Key)
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
func (s *yandexDNSProviderSolver) Initialize(config *rest.Config, _ <-chan struct{}) error {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	s.client = client
	return nil
}

// Helper function
//
// provider returns yandex.Provider configured from yandexDNSProviderConfig
// or environment variables if config not specified.
func (s *yandexDNSProviderSolver) provider(
	cfg *yandexDNSProviderConfig,
	ch *v1alpha1.ChallengeRequest,
) (*yandex.DNSProvider, error) {
	if cfg == nil {
		return yandex.NewDNSProvider(
			yandex.NewProviderConfigFromEnv(),
		)
	}

	if cfg.APIKeySecretRef.LocalObjectReference.Name == "" {
		return nil, fmt.Errorf("provider secret token or key were not provided")
	}

	secret, err := s.client.CoreV1().
		Secrets(ch.ResourceNamespace).
		Get(context.Background(), cfg.APIKeySecretRef.LocalObjectReference.Name, meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	secretData, ok := secret.Data[cfg.APIKeySecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret \"%s/%s\"",
			cfg.APIKeySecretRef.Key,
			cfg.APIKeySecretRef.LocalObjectReference.Name,
			ch.ResourceNamespace)
	}

	providerCfg := yandex.NewProviderConfig(cfg.AuthorizationType, cfg.FolderId)
	providerCfg.SetSecret(string(secretData))

	if cfg.DNSRecordSetTTL != 0 {
		providerCfg.DNSRecordSetTTL = cfg.DNSRecordSetTTL
	}

	return yandex.NewDNSProvider(providerCfg)
}

// Helper method
//
// loadConfig decodes JSON configuration into the typed config struct.
func loadConfig(cfgJSON *apiext.JSON) (*yandexDNSProviderConfig, error) {
	// handle case when empty config specified
	if cfgJSON == nil {
		return nil, nil
	}

	cfg := &yandexDNSProviderConfig{}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
