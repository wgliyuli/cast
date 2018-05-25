package cast

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"os"

	"context"

	"github.com/google/go-querystring/query"
	"github.com/jtacoma/uritemplates"
)

var defaultClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Cast struct {
	client     *http.Client
	urlPrefix  string
	api        string
	method     string
	header     http.Header
	queryParam interface{}
	pathParam  map[string]interface{}
	body       reqBody
	basicAuth  *BasicAuth
	retry      int
	strat      backoffStrategy
	retryHooks []RetryHook
	timeout    time.Duration
	logger     *log.Logger
	dumpRequestHook
	dumpResponseHook
}

func New(sl ...Setter) *Cast {
	c := new(Cast)
	c.client = defaultClient
	c.header = make(http.Header)
	c.pathParam = make(map[string]interface{})
	c.logger = log.New(os.Stderr, "CAST ", log.LstdFlags|log.Llongfile)

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

func (c *Cast) AppendHeader(header http.Header) *Cast {
	for k, vv := range header {
		for _, v := range vv {
			c.header.Add(k, v)
		}
	}
	return c
}

func (c *Cast) SetHeader(header http.Header) *Cast {
	for k, vv := range header {
		for _, v := range vv {
			c.header.Set(k, v)
		}
	}
	return c
}

func (c *Cast) WithQueryParam(queryParam interface{}) *Cast {
	c.queryParam = queryParam
	return c
}

func (c *Cast) WithPathParam(pathParam map[string]interface{}) *Cast {
	c.pathParam = pathParam
	return c
}

func (c *Cast) WithJsonBody(body interface{}) *Cast {
	c.body = reqJsonBody{
		payload: body,
	}
	return c
}

func (c *Cast) WithXmlBody(body interface{}) *Cast {
	c.body = reqXmlBody{
		payload: body,
	}
	return c
}

func (c *Cast) WithPlainBody(body string) *Cast {
	c.body = reqPlainBody{
		payload: body,
	}
	return c
}

func (c *Cast) WithUrlEncodedFormBody(body interface{}) *Cast {
	c.body = reqFormUrlEncodedBody{
		payload: body,
	}
	return c
}

func (c *Cast) WithRetry(retry int) *Cast {
	c.retry = retry
	return c
}

func (c *Cast) WithLinearBackoffStrategy(slope time.Duration) *Cast {
	c.strat = linearBackoffStrategy{
		slope: slope,
	}
	return c
}

func (c *Cast) WithConstantBackoffStrategy(internal time.Duration) *Cast {
	c.strat = constantBackOffStrategy{
		interval: internal,
	}
	return c
}

func (c *Cast) WithExponentialBackoffStrategy(base, cap time.Duration) *Cast {
	c.strat = exponentialBackoffStrategy{
		exponentialBackoff{
			base: base,
			cap:  cap,
		},
	}
	return c
}

func (c *Cast) WithExponentialBackoffEqualJitterStrategy(base, cap time.Duration) *Cast {
	c.strat = exponentialBackoffEqualJitterStrategy{
		exponentialBackoff{
			base: base,
			cap:  cap,
		},
	}
	return c
}

func (c *Cast) WithExponentialBackoffFullJitterStrategy(base, cap time.Duration) *Cast {
	c.strat = exponentialBackoffFullJitterStrategy{
		exponentialBackoff{
			base: base,
			cap:  cap,
		},
	}
	return c
}

func (c *Cast) WithExponentialBackoffDecorrelatedJitterStrategy(base, cap time.Duration) *Cast {
	c.strat = exponentialBackoffDecorrelatedJitterStrategy{
		exponentialBackoff{
			base: base,
			cap:  cap,
		},
		base,
	}
	return c
}

func (c *Cast) AddRetryHooks(hooks ...RetryHook) *Cast {
	for _, hook := range hooks {
		c.retryHooks = append(c.retryHooks, hook)
	}
	return c
}

func (c *Cast) WithTimeout(timeout time.Duration) *Cast {
	c.timeout = timeout
	return c
}

func (c *Cast) finalizeApi() error {
	if len(c.pathParam) > 0 {
		tpl, err := uritemplates.Parse(c.api)
		if err != nil {
			c.logger.Printf("ERROR [%v]", err)
			return err
		}
		c.api, err = tpl.Expand(c.pathParam)
		if err != nil {
			c.logger.Printf("ERROR [%v]", err)
			return err
		}
	}
	return nil
}

func (c *Cast) reqBody() (io.Reader, error) {
	var (
		reqBody io.Reader
		err     error
	)
	if c.body != nil {
		reqBody, err = c.body.Body()
		if err != nil {
			c.logger.Printf("ERROR [%v]", err)
			return nil, err
		}

	}

	return reqBody, nil
}

func (c *Cast) finalizeHeader(request *http.Request) {
	if c.body != nil && len(c.body.ContentType()) > 0 {
		c.SetHeader(http.Header{
			contentType: []string{c.body.ContentType()},
		})
	}

	for k, vv := range c.header {
		for _, v := range vv {
			request.Header.Add(k, v)
		}
	}
}

func (c *Cast) finalizeQueryParamIfAny(request *http.Request) error {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return err
	}

	qValues, err := query.Values(c.queryParam)
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return err
	}
	for k, vv := range qValues {
		for _, v := range vv {
			values.Add(k, v)
		}
	}
	request.URL.RawQuery = values.Encode()

	return nil
}

func (c *Cast) setBasicAuthIfAny(request *http.Request) {
	if c.basicAuth != nil {
		request.SetBasicAuth(c.basicAuth.info())
	}
}

func (c *Cast) setTimeoutIfAny(request *http.Request) {
	if c.timeout > 0 {
		ctx, cancel := context.WithCancel(context.TODO())
		_ = time.AfterFunc(c.timeout, func() {
			cancel()
		})
		request = request.WithContext(ctx)
	}
}

func (c *Cast) dumpRequestHookIfAny(request *http.Request) {
	if c.dumpRequestHook != nil {
		c.dumpRequestHook(c.logger, request)
	}
}

func (c *Cast) dumpResponseHookIfAny(response *http.Response) {
	if c.dumpResponseHook != nil {
		c.dumpResponseHook(c.logger, response)
	}
}


func (c *Cast) genReply(start time.Time, request *http.Request) (*Reply, error) {
	var (
		resp  *http.Response
		count = 0
		err   error
	)

	for {

		if count > c.retry {
			break
		}

		resp, err = c.client.Do(request)
		count++

		var isRetry bool
		for _, hook := range c.retryHooks {
			if hook(resp) != nil {
				isRetry = true
				break
			}
		}

		if (isRetry && count <= c.retry+1) || err != nil {
			if resp != nil {
				resp.Body.Close()
			}
			if c.strat != nil {
				<-time.After(c.strat.backoff(count))
				continue
			}
		}

		break
	}

	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}
	defer resp.Body.Close()

	c.dumpResponseHookIfAny(resp)

	rep := new(Reply)
	rep.statusCode = resp.StatusCode
	repBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}
	rep.body = repBody
	rep.cost = time.Since(start)
	rep.times = count

	c.logger.Printf("%s took %s upto %d time(s)", resp.Request.URL.String(), rep.cost, rep.times)

	return rep, nil
}

func (c *Cast) Request() (*Reply, error) {
	start := time.Now()

	if err := c.finalizeApi(); err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	reqBody, err := c.reqBody()
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	req, err := http.NewRequest(c.method, c.urlPrefix+c.api, reqBody)
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	c.finalizeHeader(req)

	if err := c.finalizeQueryParamIfAny(req); err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	c.setBasicAuthIfAny(req)
	c.setTimeoutIfAny(req)
	c.dumpRequestHookIfAny(req)

	rep, err := c.genReply(start, req)
	if err != nil {
		c.logger.Printf("ERROR [%v]", err)
		return nil, err
	}

	return rep, nil
}
