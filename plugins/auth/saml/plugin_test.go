package saml

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// idpMetadata is a minimal IdP metadata document with an HTTP-Redirect SSO
// binding, sufficient for building authentication requests.
const idpMetadata = `<?xml version="1.0" encoding="UTF-8"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
                  entityID="https://idp.example.com/metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</NameIDFormat>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                         Location="https://idp.example.com/sso"/>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                         Location="https://idp.example.com/sso"/>
  </IDPSSODescriptor>
</EntityDescriptor>`

func newTestLogger(t *testing.T) *observability.Logger {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return logger
}

func writeMetadataFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "idp-metadata.xml")
	require.NoError(t, os.WriteFile(path, []byte(idpMetadata), 0600))
	return path
}

func validSettings(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"entityId":        "https://sp.example.com/metadata",
		"acsUrl":          "https://sp.example.com/saml/acs",
		"idpMetadataFile": writeMetadataFile(t),
	}
}

func TestPlugin_Lifecycle(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "saml", p.Name())

	require.NoError(t, p.Initialize(validSettings(t), newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	assert.NotNil(t, p.ServiceProvider())
	assert.NoError(t, p.Stop())
}

func TestPlugin_InitializeValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]interface{})
		wantErr string
	}{
		{"valid", func(s map[string]interface{}) {}, ""},
		{"missing entityId", func(s map[string]interface{}) { delete(s, "entityId") }, "entityId"},
		{"missing acsUrl", func(s map[string]interface{}) { delete(s, "acsUrl") }, "acsUrl"},
		{"no metadata source", func(s map[string]interface{}) { delete(s, "idpMetadataFile") }, "exactly one"},
		{"both metadata sources", func(s map[string]interface{}) {
			s["idpMetadataUrl"] = "https://idp.example.com/metadata"
		}, "exactly one"},
		{"cert without key", func(s map[string]interface{}) { s["certFile"] = "cert.pem" }, "together"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := validSettings(t)
			tt.mutate(settings)

			p := &Plugin{}
			err := p.Initialize(settings, newTestLogger(t), nil, nil, nil)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestPlugin_AuthURL(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(validSettings(t), newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())

	authURL, err := p.AuthURL("my-relay-state")
	require.NoError(t, err)

	parsed, err := url.Parse(authURL)
	require.NoError(t, err)
	assert.Equal(t, "idp.example.com", parsed.Host)
	assert.Equal(t, "/sso", parsed.Path)
	assert.NotEmpty(t, parsed.Query().Get("SAMLRequest"))
	assert.Equal(t, "my-relay-state", parsed.Query().Get("RelayState"))
}

func TestPlugin_AuthURLBeforeStart(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(validSettings(t), newTestLogger(t), nil, nil, nil))

	_, err := p.AuthURL("state")
	assert.ErrorContains(t, err, "not started")
}

func TestPlugin_SPMetadata(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(validSettings(t), newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())

	metadata, err := p.SPMetadata()
	require.NoError(t, err)
	assert.Contains(t, string(metadata), "https://sp.example.com/metadata")
	assert.Contains(t, string(metadata), "https://sp.example.com/saml/acs")
}

func TestPlugin_MetadataFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/samlmetadata+xml")
		_, _ = w.Write([]byte(idpMetadata))
	}))
	defer server.Close()

	settings := map[string]interface{}{
		"entityId":       "https://sp.example.com/metadata",
		"acsUrl":         "https://sp.example.com/saml/acs",
		"idpMetadataUrl": server.URL,
	}

	p := &Plugin{}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	assert.Equal(t, "https://idp.example.com/metadata", p.ServiceProvider().IDPMetadata.EntityID)
}

func TestPlugin_StartFailsOnMissingMetadataFile(t *testing.T) {
	settings := map[string]interface{}{
		"entityId":        "https://sp.example.com/metadata",
		"acsUrl":          "https://sp.example.com/saml/acs",
		"idpMetadataFile": filepath.Join(t.TempDir(), "missing.xml"),
	}

	p := &Plugin{}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))
	assert.Error(t, p.Start())
}
