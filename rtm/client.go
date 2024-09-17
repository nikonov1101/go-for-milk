package rtm

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Client struct {
	apiKey       string
	sharedSecret string
	authToken    string
}

func New(key, shared string) (Client, error) {
	cli := Client{apiKey: key, sharedSecret: shared}

	// try to load the cached authToken from $HOME/.rtm-token
	authToken, err := loadCachedToken()
	if err != nil {
		debugf("unable to load cached token... let's re-issue one.")

		// if there is no token (fresh start?), then go issue one.
		// authFull will print the URL you have to follow and authorize the app.
		authToken, err = cli.authFull()
		if err != nil {
			debugf("failed to perform full auth: %v", err)
			return Client{}, err
		}

		if err := saveCahcedToken(authToken); err != nil {
			debugf("failed to save cached token: %v", err)
			return Client{}, err
		}
	}

	if err := cli.checkAuthToken(authToken); err != nil {
		return Client{}, errors.Wrap(err, "check auth token")
	}

	cli.authToken = authToken
	return cli, nil
}

func (c Client) AddTask(name string) error {
	vs := url.Values{}
	vs.Add("method", methodAddTask)
	vs.Add("api_key", c.apiKey)
	vs.Add("auth_token", c.authToken)
	vs.Add("timeline", "1")
	vs.Add("name", name)
	vs.Add("parse", "1")

	target := c.signedURL(vs)
	_, err := httpGet(target)
	return err
}

func (c Client) ListTasks() ([]Task, error) {
	vs := url.Values{}
	vs.Add("method", methodListTasks)
	vs.Add("api_key", c.apiKey)
	vs.Add("auth_token", c.authToken)

	target := c.signedURL(vs)
	bs, err := httpGet(target)
	if err != nil {
		return nil, err
	}

	list := listTasksResponse{}
	if err := xml.Unmarshal(bs, &list); err != nil {
		return nil, errors.Wrap(err, "parse XML response")
	}

	tasks := list.intoTasks()
	return tasks, nil
}

func (c Client) signedURL(vs url.Values) string {
	sig := signature(vs, c.sharedSecret)
	vs.Add("api_sig", sig)
	target := apiURL + "?" + vs.Encode()

	return target
}

func signature(values url.Values, sharedSecret string) string {
	keys := []string{}
	for k, v := range values {
		keys = append(keys, fmt.Sprintf("%s%s", k, v[0]))
	}

	sort.Strings(keys)
	raw := sharedSecret + strings.Join(keys, "")
	hash := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

func httpGet(target string) ([]byte, error) {
	debugf("http request: %q", target)
	resp, err := http.DefaultClient.Get(target)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("http request fail with non-200 status code: " + resp.Status)
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
