package ldap

import (
	"fmt"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeConn is an in-memory directory with one service account and one user.
type fakeConn struct {
	serviceDN   string
	servicePass string
	userDN      string
	userPass    string
	entries     []*goldap.Entry
	searchErr   error
	closed      bool
}

func (f *fakeConn) Bind(username, password string) error {
	if username == f.serviceDN && password == f.servicePass {
		return nil
	}
	if username == f.userDN && password == f.userPass {
		return nil
	}
	return fmt.Errorf("invalid credentials for %s", username)
}

func (f *fakeConn) Search(req *goldap.SearchRequest) (*goldap.SearchResult, error) {
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return &goldap.SearchResult{Entries: f.entries}, nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func validSettings() map[string]interface{} {
	return map[string]interface{}{
		"url":          "ldap://ldap.example.com:389",
		"bindDn":       "cn=service,dc=example,dc=com",
		"bindPassword": "service-secret",
		"baseDn":       "ou=people,dc=example,dc=com",
	}
}

func newTestPlugin(t *testing.T, conn *fakeConn) *Plugin {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)

	p := &Plugin{dialer: func(url string, insecure bool) (Connection, error) {
		return conn, nil
	}}
	require.NoError(t, p.Initialize(validSettings(), logger, nil, nil, nil))
	return p
}

func userEntry() *goldap.Entry {
	return &goldap.Entry{
		DN: "uid=jdoe,ou=people,dc=example,dc=com",
		Attributes: []*goldap.EntryAttribute{
			{Name: "mail", Values: []string{"jdoe@example.com"}},
			{Name: "memberOf", Values: []string{"cn=devs", "cn=admins"}},
		},
	}
}

func workingConn() *fakeConn {
	return &fakeConn{
		serviceDN:   "cn=service,dc=example,dc=com",
		servicePass: "service-secret",
		userDN:      "uid=jdoe,ou=people,dc=example,dc=com",
		userPass:    "user-secret",
		entries:     []*goldap.Entry{userEntry()},
	}
}

func TestPlugin_Lifecycle(t *testing.T) {
	p := newTestPlugin(t, workingConn())

	assert.Equal(t, "ldap", p.Name())
	assert.NoError(t, p.Start())
	assert.NoError(t, p.Stop())
}

func TestPlugin_InitializeValidation(t *testing.T) {
	tests := []struct {
		name    string
		drop    string
		wantErr bool
	}{
		{"all required present", "", false},
		{"missing url", "url", true},
		{"missing bindDn", "bindDn", true},
		{"missing bindPassword", "bindPassword", true},
		{"missing baseDn", "baseDn", true},
	}

	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := validSettings()
			if tt.drop != "" {
				delete(settings, tt.drop)
			}
			p := &Plugin{}
			err := p.Initialize(settings, logger, nil, nil, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPlugin_Authenticate(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		conn := workingConn()
		p := newTestPlugin(t, conn)

		user, err := p.Authenticate("jdoe", "user-secret")
		require.NoError(t, err)
		assert.Equal(t, "jdoe", user.Username)
		assert.Equal(t, "uid=jdoe,ou=people,dc=example,dc=com", user.DN)
		assert.Equal(t, "jdoe@example.com", user.Email)
		assert.Equal(t, []string{"cn=devs", "cn=admins"}, user.Groups)
		assert.True(t, conn.closed, "connection must be closed")
	})

	t.Run("wrong password", func(t *testing.T) {
		p := newTestPlugin(t, workingConn())
		_, err := p.Authenticate("jdoe", "wrong")
		assert.ErrorContains(t, err, "invalid credentials")
	})

	t.Run("unknown user", func(t *testing.T) {
		conn := workingConn()
		conn.entries = nil
		p := newTestPlugin(t, conn)
		_, err := p.Authenticate("ghost", "whatever")
		assert.ErrorContains(t, err, "user not found")
	})

	t.Run("service bind failure", func(t *testing.T) {
		conn := workingConn()
		conn.servicePass = "rotated"
		p := newTestPlugin(t, conn)
		_, err := p.Authenticate("jdoe", "user-secret")
		assert.ErrorContains(t, err, "service bind failed")
	})

	t.Run("search failure", func(t *testing.T) {
		conn := workingConn()
		conn.searchErr = fmt.Errorf("directory unavailable")
		p := newTestPlugin(t, conn)
		_, err := p.Authenticate("jdoe", "user-secret")
		assert.ErrorContains(t, err, "user search failed")
	})

	t.Run("empty credentials", func(t *testing.T) {
		p := newTestPlugin(t, workingConn())
		_, err := p.Authenticate("", "")
		assert.Error(t, err)
	})
}

func TestPlugin_Ping(t *testing.T) {
	p := newTestPlugin(t, workingConn())
	assert.NoError(t, p.ping())

	bad := workingConn()
	bad.servicePass = "rotated"
	p2 := newTestPlugin(t, bad)
	assert.Error(t, p2.ping())
}
