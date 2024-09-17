package rtm

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// do full auth, normally needed only once to create and cache the auth token.
func (c Client) authFull() (string, error) {
	frob, err := c.generateNewFrob()
	if err != nil {
		return "", errors.Wrap(err, "generate frob")
	}

	tmpAuthURL := c.makeDesktopAuthURL(frob)
	fmt.Printf("open auth URL in your browser:\n%s\n", tmpAuthURL)
	<-time.After(2 * time.Second)

	fmt.Printf("press any key to continue authorization process...")
	os.Stdin.Read(make([]byte, 1))

	// exchange the short-living frob for the long-living auth token
	token, err := c.getAuthToken(frob)
	if err != nil {
		return "", errors.Wrap(err, "get auth token")
	}

	return token, nil
}

func (c Client) generateNewFrob() (string, error) {
	vs := url.Values{}
	vs.Add("method", methodGetFrom)
	vs.Add("api_key", c.apiKey)
	vs.Add("perms", "write")

	target := c.signedURL(vs)
	bs, err := httpGet(target)
	if err != nil {
		return "", err
	}
	var frob frobResponse
	if err := xml.Unmarshal(bs, &frob); err != nil {
		return "", errors.Wrap(err, "parse response XML")
	}

	debugf("grob generated: %q", frob.Frob)
	return frob.Frob, nil
}

func (c Client) makeDesktopAuthURL(frob string) string {
	vs := url.Values{}
	vs.Add("api_key", c.apiKey)
	vs.Add("perms", "write")
	vs.Add("frob", frob)

	sig := signature(vs, c.sharedSecret)
	vs.Add("api_sig", sig)
	// note: do not use `signedURL` here because we need the authURL
	// (to be displayed in a browser) rather than apiURL.
	target := authURL + "?" + vs.Encode()
	return target
}

func (c Client) getAuthToken(frob string) (string, error) {
	vs := url.Values{}
	vs.Add("api_key", c.apiKey)
	vs.Add("method", methodGetAuthToken)
	vs.Add("frob", frob)

	target := c.signedURL(vs)
	bs, err := httpGet(target)
	if err != nil {
		return "", err
	}

	debugf("get auth token: response: %s", string(bs))
	tokenResp := authTokenResponse{}
	if err := xml.Unmarshal(bs, &tokenResp); err != nil {
		return "", errors.Wrap(err, "parse XML response")
	}

	if tokenResp.Stat == "fail" {
		return "", errors.New("token request failed: have you visited the authorization URL?")
	}

	debugf("authorized as @%s (%s)", tokenResp.Auth.User.Username, tokenResp.Auth.User.Fullname)
	return tokenResp.Auth.Token, nil
}

func (c Client) checkAuthToken(token string) error {
	vs := url.Values{}
	vs.Add("method", methodCheckAuthToken)
	vs.Add("api_key", c.apiKey)
	vs.Add("auth_token", token)

	target := c.signedURL(vs)
	if bs, err := httpGet(target); err != nil {
		// be more verbose on why auth token is not valid
		debugf("check auth token: FAILED: %v", string(bs))
		return err
	}

	debugf("check auth token: OK")
	return nil
}

func loadCachedToken() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "get user's home dir")
	}

	tokenCachedPath := filepath.Join(homedir, ".rtm-token")
	debugf("loading cached token from %q...", tokenCachedPath)

	bs, err := os.ReadFile(tokenCachedPath)
	if err != nil {
		return "", errors.Wrap(err, "read cached token")
	}
	ct := cachedToken{}
	if err := json.Unmarshal(bs, &ct); err != nil {
		return "", errors.Wrap(err, "unmarshal cached token JSON")
	}

	debugf("cached token loaded: updated %s (%s ago)", ct.UpdatedAt, time.Since(ct.UpdatedAt))
	return ct.Token, nil
}

func saveCahcedToken(token string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user's home dir")
	}

	tokenCachedPath := filepath.Join(homedir, ".rtm-token")
	debugf("saving cached token to %q...", tokenCachedPath)

	ct := cachedToken{
		Token:     token,
		UpdatedAt: time.Now().UTC(),
	}

	bs, _ := json.Marshal(ct)
	if err := os.WriteFile(tokenCachedPath, bs, 0o600); err != nil {
		errors.Wrap(err, "write token cache file")
	}

	debugf("cached token saved to %q", tokenCachedPath)
	return nil
}

func debugf(f string, args ...any) {
	// WARN: will leak keys
	if v := os.Getenv("RTM_DEBUG"); len(v) > 0 {
		log.Printf(f, args...)
	}
}
