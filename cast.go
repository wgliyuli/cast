package cast

import (
	"net/http"
	"net/url"
	"time"
	"io/ioutil"
	"io"
)

var defaultClient = &http.Client{
	Timeout: 3 * time.Second,
}

type Cast struct {
	client *http.Client
	api string
	method string
	header http.Header
	queryParam url.Values
	pathParam url.Values
	body ReqBody
	basicAuth *BasicAuth
}

func New(sl ...Setter) *Cast {
	c := new(Cast)
	c.client = defaultClient

	for _, s := range sl {
		s(c)
	}

	return c
}

func (c *Cast) WithApi(api string) *Cast {
	c.api = api
	return c
}

func (c *Cast) WithMethod(method string) *Cast {
		c.method = method
		return c
}

func (c *Cast) WithHeader(header http.Header) *Cast {
	c.header = header
	return c
}

func (c * Cast) WithQueryParam(queryParam url.Values) *Cast {
	c.queryParam = queryParam
	return c
}

func (c * Cast) WithPathParam(pathParam url.Values) *Cast {
	c.pathParam = pathParam
	return c
}

func (c *Cast) WithBody(body ReqBody) *Cast {
	c.body = body
	return c
}

func (c *Cast) Request() (*Reply, error) {
	var (
		reqBody io.Reader
		err error
	)
	if c.body != nil {
		reqBody, err = c.body.Body()
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(c.method, c.api, reqBody)
	if err != nil {
		return nil, err
	}

	for k, vv := range c.header {
		for _, v := range(vv) {
			req.Header.Add(k, v)
		}
	}

	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	for k, vv := range c.queryParam {
		for _, v := range(vv) {
			values.Add(k, v)
		}
	}
	req.URL.RawQuery = values.Encode()

	if c.basicAuth != nil {
		req.SetBasicAuth(c.basicAuth.Info())
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rep := new(Reply)
	rep.statusCode = resp.StatusCode
	repBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rep.body = repBody
	c.Reset()
	return rep, nil
}

func (c *Cast) Reset() {
	copy := New()
	copy.client = c.client
	copy.basicAuth = c.basicAuth
	c = copy
	return
}
