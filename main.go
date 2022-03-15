package main

import (
	"context"
	"encoding/json"
	"errors"
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

// customDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
type yandexDNSProviderConfig struct {
	APIKeySecretRef core.SecretKeySelector `json:"apiKeySecretRef"`
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

	provider, err := s.provider(&cfg, ch.ResourceNamespace)
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

	provider, err := s.provider(&cfg, ch.ResourceNamespace)
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

// validate is a helper function that validates provider config
func (c *yandexDNSProviderConfig) validate() error {
	if c.APIKeySecretRef.LocalObjectReference.Name == "" {
		return errors.New("API token field were not provided")
	}

	return nil
}

// provider is a helper function that creates provider from config
//
// returns YandexProvider or error if any error occurred
func (s *yandexDNSProviderSolver) provider(c *yandexDNSProviderConfig, namespace string) (*yandex.DNSProvider, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	secret, err := s.client.CoreV1().
		Secrets(namespace).
		Get(context.Background(), c.APIKeySecretRef.LocalObjectReference.Name, meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	// todo: use secret state as provider api-key or authorized key
	_, ok := secret.Data[c.APIKeySecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret \"%s/%s\"",
			c.APIKeySecretRef.Key,
			c.APIKeySecretRef.LocalObjectReference.Name,
			namespace)
	}

	provider := yandex.NewProvider(
		&yandex.DNSProviderConfig{},
	)

	return provider, nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *apiext.JSON) (yandexDNSProviderConfig, error) {
	cfg := yandexDNSProviderConfig{}

	if cfgJSON == nil {
		return cfg, nil
	}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
