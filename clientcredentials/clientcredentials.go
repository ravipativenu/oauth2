package cliientcredentials // import "golang.org/x/oauth2/clientcredentials"
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"venuravipati/oauth2"
	"venuravipati/oauth2/internal"
)

// Config describes a 2-legged OAuth2 flow, with both the
// client application information and the server's endpoint URLs.
type Config struct {

	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// TokenURL is the resource server's token endpoint
	// URL. This is a constant specific to each server.
	TokenURL string

	// Scope specifies optional requested permissions.
	Scopes []string

	// EndpointParams specifies additional parameters for requests to the token endpoint.
	EndpointParams url.Values

	// AuthStyle optionally specifies how the endpoint wants the
	// client ID & client secret sent. The zero value means to
	// auto-detect.
	AuthStyle oauth2.AuthStyle
}

// Token uses client credentials to retrieve a token.
//
// The provided context optionally controls which HTTP client is used. See the oauth2.HTTPClient variable.
func (c *Config) Token(ctx context.Context) (*oauth2.Token, error) {
	return c.TokenSource(ctx).Token()
}

// Client returns an HTTP client using the provided token.
// The token will auto-refresh as necessary.
//
// The provided context optionally controls which HTTP client
// is returned. See the oauth2.HTTPClient variable.
//
// The returned Client and its Transport should not be modified.
func (c *Config) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx))
}

// TokenSource returns a TokenSource that returns t until t expires,
// automatically refreshing it as necessary using the provided context and the
// client ID and client secret.
//
// Most users will use Config.Client instead.
func (c *Config) TokenSource(ctx context.Context) oauth2.TokenSource {
	fmt.Println("Inside Config.TokenSource...")
	source := &tokenSource{
		ctx:  ctx,
		conf: c,
	}
	fmt.Println(*source.conf)
	return oauth2.ReuseTokenSource(nil, source)
}

type tokenSource struct {
	ctx  context.Context
	conf *Config
}

// Token refreshes the token by using a new client credentials request.
// tokens received this way do not include a refresh token
func (c *tokenSource) Token() (*oauth2.Token, error) {
	v := url.Values{
		"grant_type": {"client_credentials"},
	}
	if len(c.conf.Scopes) > 0 {
		v.Set("scope", strings.Join(c.conf.Scopes, " "))
	}
	for k, p := range c.conf.EndpointParams {
		// Allow grant_type to be overridden to allow interoperability with
		// non-compliant implementations.
		if _, ok := v[k]; ok && k != "grant_type" {
			return nil, fmt.Errorf("oauth2: cannot overwrite parameter %q", k)
		}
		v[k] = p
	}

	tk, err := internal.RetrieveToken(c.ctx, c.conf.ClientID, c.conf.ClientSecret, c.conf.TokenURL, v, internal.AuthStyle(c.conf.AuthStyle))
	if err != nil {
		if rErr, ok := err.(*internal.RetrieveError); ok {
			return nil, (*oauth2.RetrieveError)(rErr)
		}
		return nil, err
	}
	t := &oauth2.Token{
		AccessToken:  tk.AccessToken,
		TokenType:    tk.TokenType,
		RefreshToken: tk.RefreshToken,
		Expiry:       tk.Expiry,
	}
	return t.WithExtra(tk.Raw), nil
}
