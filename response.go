package cast

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Response wraps the raw response with attributes.
type Response struct {
	request     *Request
	rawResponse *http.Response
	statusCode  int
	body        []byte
}

// StatusCode returns http status code.
func (resp *Response) StatusCode() int {
	return resp.statusCode
}

// Cookies returns http cookies.
func (resp *Response) Cookies() []*http.Cookie {
	if resp.rawResponse == nil {
		return []*http.Cookie{}
	}
	return resp.rawResponse.Cookies()
}

// Body returns the underlying response body.
func (resp *Response) Body() []byte {
	return resp.body
}

// String returns the underlying body in string.
func (resp *Response) String() string {
	return string(resp.body)
}

// DecodeFromJSON decodes the JSON body into data variable.
func (resp *Response) DecodeFromJSON(v interface{}) error {
	if len(resp.body) == 0 {
		return nil
	}
	return json.Unmarshal(resp.body, &v)
}

// DecodeFromXML decodes the XML body into  data variable.
func (resp *Response) DecodeFromXML(v interface{}) error {
	if len(resp.body) == 0 {
		return nil
	}
	return xml.Unmarshal(resp.body, &v)
}

// Size returns the length of the body.
func (resp *Response) Size() int64 {
	if resp.rawResponse == nil {
		return 0
	}
	return resp.rawResponse.ContentLength
}

// Header returns the response header.
func (resp *Response) Header() http.Header {
	if resp.rawResponse == nil {
		return http.Header{}
	}
	return resp.rawResponse.Header
}

// StatusOk returns true if http status code is 200, otherwise false.
func (resp *Response) StatusOk() bool {
	return resp.statusCode == http.StatusOK
}

// Success returns true if http status code is in [200,299], otherwise false.
func (resp *Response) Success() bool {
	return resp.statusCode <= 299 && resp.statusCode >= 200
}

// Method returns the request method.
func (resp *Response) Method() string {
	if resp == nil {
		return ""
	}
	if resp.request == nil {
		return ""
	}
	if resp.request.rawRequest == nil {
		return ""
	}
	return resp.request.rawRequest.Method
}

// URL returns the request url.
func (resp *Response) URL() string {
	if resp == nil {
		return ""
	}
	if resp.request == nil {
		return ""
	}
	if resp.request.rawRequest == nil {
		return ""
	}
	return resp.request.rawRequest.URL.String()
}

// SetHeader sets the key, value pair list.
func (resp *Response) SetHeader(vv ...string) *Response {
	if resp == nil {
		return nil
	}
	if resp.request == nil {
		return nil
	}
	if resp.request.rawRequest == nil {
		return nil
	}
	if len(vv)%2 != 0 {
		return nil
	}
	for i := 0; i < len(vv); i += 2 {
		resp.request.rawRequest.Header.Set(vv[i], vv[i+1])
	}
	return resp
}

// AddHeader adds the key, value pair list.
func (resp *Response) AddHeader(vv ...string) *Response {
	if resp == nil {
		return nil
	}
	if resp.request == nil {
		return nil
	}
	if resp.request.rawRequest == nil {
		return nil
	}
	if len(vv)%2 != 0 {
		return nil
	}
	for i := 0; i < len(vv); i += 2 {
		resp.request.rawRequest.Header.Add(vv[i], vv[i+1])
	}
	return resp
}
