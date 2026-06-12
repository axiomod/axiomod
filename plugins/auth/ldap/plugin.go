// Package ldap provides a plugin that authenticates users against an LDAP
// directory using the bind-search-bind flow.
package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/plugins"

	goldap "github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"
)

const defaultUserFilter = "(uid=%s)"

// Connection is the subset of the LDAP connection used by the plugin. It
// allows tests to substitute a fake directory.
type Connection interface {
	Bind(username, password string) error
	Search(req *goldap.SearchRequest) (*goldap.SearchResult, error)
	Close() error
}

// Dialer opens a connection to the directory.
type Dialer func(url string, insecureSkipVerify bool) (Connection, error)

// defaultDialer connects with go-ldap, using TLS settings appropriate for
// the scheme in the URL (ldap:// or ldaps://).
func defaultDialer(url string, insecureSkipVerify bool) (Connection, error) {
	conn, err := goldap.DialURL(url, goldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // explicit opt-in for dev environments
	}))
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(10 * time.Second)
	return conn, nil
}

// User is the directory entry resolved during authentication.
type User struct {
	DN       string
	Username string
	Email    string
	Groups   []string
}

// Plugin authenticates users against an LDAP directory.
type Plugin struct {
	logger *observability.Logger
	dialer Dialer

	url                string
	bindDN             string
	bindPassword       string
	baseDN             string
	userFilter         string
	emailAttribute     string
	groupAttribute     string
	insecureSkipVerify bool
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "ldap"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - url (string): directory URL, e.g. "ldaps://ldap.example.com:636" (required)
//   - bindDn (string): service account DN used for the search bind (required)
//   - bindPassword (string): service account password (required)
//   - baseDn (string): search base, e.g. "ou=people,dc=example,dc=com" (required)
//   - userFilter (string): search filter with one %s for the username, default "(uid=%s)"
//   - emailAttribute (string): attribute holding the user's email, default "mail"
//   - groupAttribute (string): attribute holding group memberships, default "memberOf"
//   - insecureSkipVerify (bool): skip TLS verification (dev only), default false
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, h *health.Health) error {
	settings = plugins.NormalizeSettings(settings)
	p.logger = logger
	if p.dialer == nil {
		p.dialer = defaultDialer
	}
	p.userFilter = defaultUserFilter
	p.emailAttribute = "mail"
	p.groupAttribute = "memberOf"

	p.url, _ = settings["url"].(string)
	p.bindDN, _ = settings["binddn"].(string)
	p.bindPassword, _ = settings["bindpassword"].(string)
	p.baseDN, _ = settings["basedn"].(string)

	if p.url == "" || p.bindDN == "" || p.bindPassword == "" || p.baseDN == "" {
		return fmt.Errorf("ldap plugin: url, bindDn, bindPassword and baseDn settings are required")
	}

	if filter, ok := settings["userfilter"].(string); ok && filter != "" {
		p.userFilter = filter
	}
	if attr, ok := settings["emailattribute"].(string); ok && attr != "" {
		p.emailAttribute = attr
	}
	if attr, ok := settings["groupattribute"].(string); ok && attr != "" {
		p.groupAttribute = attr
	}
	if skip, ok := settings["insecureskipverify"].(bool); ok {
		p.insecureSkipVerify = skip
	}

	if h != nil {
		h.RegisterCheck(p.Name(), p.ping)
	}

	return nil
}

// Start starts the plugin.
func (p *Plugin) Start() error {
	if p.logger != nil {
		p.logger.Info("LDAP plugin started", zap.String("url", p.url), zap.String("base_dn", p.baseDN))
	}
	return nil
}

// Stop stops the plugin.
func (p *Plugin) Stop() error {
	return nil
}

// Authenticate verifies the user's credentials with the bind-search-bind
// flow: bind as the service account, search for the user's DN, then bind as
// the user with the supplied password.
func (p *Plugin) Authenticate(username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("ldap plugin: username and password are required")
	}

	conn, err := p.dialer(p.url, p.insecureSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("ldap plugin: failed to connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.Bind(p.bindDN, p.bindPassword); err != nil {
		return nil, fmt.Errorf("ldap plugin: service bind failed: %w", err)
	}

	searchReq := goldap.NewSearchRequest(
		p.baseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		1, 0, false,
		fmt.Sprintf(p.userFilter, goldap.EscapeFilter(username)),
		[]string{"dn", p.emailAttribute, p.groupAttribute},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("ldap plugin: user search failed: %w", err)
	}
	if len(result.Entries) != 1 {
		return nil, fmt.Errorf("ldap plugin: user not found")
	}

	entry := result.Entries[0]
	if err := conn.Bind(entry.DN, password); err != nil {
		return nil, fmt.Errorf("ldap plugin: invalid credentials")
	}

	return &User{
		DN:       entry.DN,
		Username: username,
		Email:    entry.GetAttributeValue(p.emailAttribute),
		Groups:   entry.GetAttributeValues(p.groupAttribute),
	}, nil
}

// ping verifies the directory is reachable and the service bind works.
func (p *Plugin) ping() error {
	conn, err := p.dialer(p.url, p.insecureSkipVerify)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	return conn.Bind(p.bindDN, p.bindPassword)
}
