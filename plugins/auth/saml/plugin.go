// Package saml provides a plugin that configures a SAML 2.0 service
// provider: it loads IdP metadata, exposes SP metadata, and produces
// authentication request URLs for the HTTP-Redirect binding.
package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	crewsaml "github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"go.uber.org/zap"
)

// Plugin implements a SAML 2.0 service provider.
type Plugin struct {
	logger *observability.Logger

	entityID        string
	acsURL          string
	idpMetadataURL  string
	idpMetadataFile string
	certFile        string
	keyFile         string

	sp *crewsaml.ServiceProvider
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "saml"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - entityId (string): SP entity ID (required)
//   - acsUrl (string): assertion consumer service URL (required)
//   - idpMetadataUrl (string): URL of the IdP metadata document
//   - idpMetadataFile (string): path of a local IdP metadata document
//     (exactly one of idpMetadataUrl / idpMetadataFile is required)
//   - certFile (string): SP certificate (PEM), optional, enables signing
//   - keyFile (string): SP private key (PEM), optional, enables signing
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, h *health.Health) error {
	p.logger = logger

	p.entityID, _ = settings["entityId"].(string)
	p.acsURL, _ = settings["acsUrl"].(string)
	p.idpMetadataURL, _ = settings["idpMetadataUrl"].(string)
	p.idpMetadataFile, _ = settings["idpMetadataFile"].(string)
	p.certFile, _ = settings["certFile"].(string)
	p.keyFile, _ = settings["keyFile"].(string)

	if p.entityID == "" || p.acsURL == "" {
		return fmt.Errorf("saml plugin: entityId and acsUrl settings are required")
	}
	if (p.idpMetadataURL == "") == (p.idpMetadataFile == "") {
		return fmt.Errorf("saml plugin: exactly one of idpMetadataUrl or idpMetadataFile is required")
	}
	if (p.certFile == "") != (p.keyFile == "") {
		return fmt.Errorf("saml plugin: certFile and keyFile must be set together")
	}

	return nil
}

// Start loads the IdP metadata and builds the service provider.
func (p *Plugin) Start() error {
	idpMetadata, err := p.loadIDPMetadata()
	if err != nil {
		return err
	}

	acsURL, err := url.Parse(p.acsURL)
	if err != nil {
		return fmt.Errorf("saml plugin: invalid acsUrl: %w", err)
	}
	entityURL, err := url.Parse(p.entityID)
	if err != nil {
		return fmt.Errorf("saml plugin: invalid entityId: %w", err)
	}

	sp := &crewsaml.ServiceProvider{
		EntityID:    p.entityID,
		AcsURL:      *acsURL,
		MetadataURL: *entityURL,
		IDPMetadata: idpMetadata,
	}

	if p.certFile != "" {
		keyPair, err := tls.LoadX509KeyPair(p.certFile, p.keyFile)
		if err != nil {
			return fmt.Errorf("saml plugin: failed to load SP key pair: %w", err)
		}
		cert, err := x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return fmt.Errorf("saml plugin: failed to parse SP certificate: %w", err)
		}
		key, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("saml plugin: SP private key must be RSA")
		}
		sp.Key = key
		sp.Certificate = cert
	}

	p.sp = sp

	if p.logger != nil {
		p.logger.Info("SAML plugin started",
			zap.String("entity_id", p.entityID),
			zap.String("acs_url", p.acsURL),
			zap.String("idp_entity_id", idpMetadata.EntityID),
		)
	}
	return nil
}

// Stop stops the plugin.
func (p *Plugin) Stop() error {
	return nil
}

// ServiceProvider returns the underlying SAML service provider, or nil
// before Start.
func (p *Plugin) ServiceProvider() *crewsaml.ServiceProvider {
	return p.sp
}

// AuthURL builds an authentication request URL for the HTTP-Redirect
// binding, carrying the given relay state.
func (p *Plugin) AuthURL(relayState string) (string, error) {
	if p.sp == nil {
		return "", fmt.Errorf("saml plugin: not started")
	}

	authURL, err := p.sp.MakeRedirectAuthenticationRequest(relayState)
	if err != nil {
		return "", fmt.Errorf("saml plugin: failed to build authentication request: %w", err)
	}
	return authURL.String(), nil
}

// SPMetadata returns the service provider metadata document as XML.
func (p *Plugin) SPMetadata() ([]byte, error) {
	if p.sp == nil {
		return nil, fmt.Errorf("saml plugin: not started")
	}
	return xml.MarshalIndent(p.sp.Metadata(), "", "  ")
}

// loadIDPMetadata reads the IdP metadata from the configured file or URL.
func (p *Plugin) loadIDPMetadata() (*crewsaml.EntityDescriptor, error) {
	if p.idpMetadataFile != "" {
		data, err := os.ReadFile(p.idpMetadataFile)
		if err != nil {
			return nil, fmt.Errorf("saml plugin: failed to read IdP metadata file: %w", err)
		}
		metadata, err := samlsp.ParseMetadata(data)
		if err != nil {
			return nil, fmt.Errorf("saml plugin: failed to parse IdP metadata: %w", err)
		}
		return metadata, nil
	}

	metadataURL, err := url.Parse(p.idpMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("saml plugin: invalid idpMetadataUrl: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metadata, err := samlsp.FetchMetadata(ctx, http.DefaultClient, *metadataURL)
	if err != nil {
		return nil, fmt.Errorf("saml plugin: failed to fetch IdP metadata: %w", err)
	}
	return metadata, nil
}
