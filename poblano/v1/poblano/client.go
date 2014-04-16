/*
   Copyright (C) 2013  Salsita s.r.o.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program. If not, see {http://www.gnu.org/licenses/}.
*/

package poblano

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	LibraryVersion = "0.0.1"

	defaultUserAgent = "go-poblano/" + LibraryVersion
)

type Client struct {
	// Poblano access token to be used to authenticate API requests.
	token string

	// Optional Basic auth credentials to use when connecting to Poblano.
	credentials *Credentials

	// HTTP client to be used for communication with the Poblano API.
	client *http.Client

	// Base URL of the Poblano API that is to be used to form API requests.
	baseURL *url.URL

	// User-Agent header to use when connecting to the Poblano API.
	UserAgent string

	// GitHub service encapsulates all the functionality connected to GitHub.
	GitHub *GitHubService
}

type Credentials struct {
	Username string
	Password string
}

func NewClient(baseURL, apiToken string, cred *Credentials) (*Client, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		token:       apiToken,
		credentials: cred,
		client:      http.DefaultClient,
		baseURL:     base,
		UserAgent:   defaultUserAgent,
	}
	c.GitHub = newGitHubService(c)

	return c, nil
}

func (c *Client) NewRequest(method, urlPath string, body interface{}) (*http.Request, error) {
	relativePath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(relativePath)

	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if cred := c.credentials; cred != nil {
		req.SetBasicAuth(cred.Username, cred.Password)
	}

	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("X-PoblanoToken", c.token)
	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return resp, &ErrHTTP{resp}
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}
