package client

// Client package inspired by google github client https://github.com/google/go-github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"code.google.com/p/go.net/context"
	"github.com/facebookgo/stackerr"
	"github.com/google/go-querystring/query"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

const (
	libraryVersion = "0.1"
	userAgent      = "go-bearded-client/" + libraryVersion
	defaultBaseURL = "http://127.0.0.1:3003/api/"
	mediaTypeV1    = "application/json"
	apiVersion     = 1
)

// A Client manages communication with the Bearded API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Bearded API.
	UserAgent string

	// Show different debug information
	Debug bool

	// Services used for talking to different parts of the Bearded API.
	Plugins *PluginsService
	Plans   *PlansService
	Agents  *AgentsService
}

// NewClient returns a new Bearded API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the goauth2 library).
func NewClient(baseUrl string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if baseUrl == "" {
		baseUrl = defaultBaseURL
	}
	baseURL, _ := url.Parse(baseUrl)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.Plugins = &PluginsService{client: c}
	c.Plans = &PlansService{client: c}
	c.Agents = &AgentsService{client: c}
	return c
}

// addOptions adds the parameters in opt as URL query parameters to s.  opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (c *Client) SetBaseUrl(u string) error {
	baseURL, err := url.Parse(u)
	if err != nil {
		return err
	}
	c.BaseURL = baseURL
	return nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.  If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	urlStr = fmt.Sprintf("v%d/%s", apiVersion, urlStr)
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	if c.Debug {
		logrus.Debugf("%s %s", method, u.String())
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", mediaTypeV1)
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		req.Header.Add("Content-Type", "application/json")
	}
	return req, nil
}

// Do sends an API request and returns the API response.  The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	var resp *http.Response
	ret := make(chan error, 1)
	go func() {
		var err error
		resp, err = c.client.Do(req)
		ret <- err
	}()
	select {
	case <-ctx.Done():
		type canceler interface {
			CancelRequest(*http.Request)
		}
		transport := c.client.Transport
		if transport == nil {
			// default transport is used
			transport = http.DefaultTransport

		}
		tr, ok := transport.(canceler)
		if !ok {
			return nil, fmt.Errorf("client Transport of type %T doesn't support CancelRequest; Timeout not supported", transport)
		}
		tr.CancelRequest(req)
		<-ret // Wait goroutine to return after cancellation.
		return nil, stackerr.Wrap(ctx.Err())
	case err := <-ret:
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	err := CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	return resp, err
}

// Helper method to get a list of payload objects
func (c *Client) List(ctx context.Context, url string, opts interface{}, payload interface{}) error {
	u, err := addOptions(url, opts)
	if err != nil {
		return err
	}

	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(ctx, req, payload)
	return err
}

// Helper method to get a resource by id
func (c *Client) Get(ctx context.Context, url string, id string, payload interface{}) error {
	url = fmt.Sprintf("%s/%s", url, id)
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(ctx, req, payload)
	return err
}

func (c *Client) Create(ctx context.Context, url string, send interface{}, payload interface{}) error {
	req, err := c.NewRequest("POST", url, send)
	if err != nil {
		return err
	}
	_, err = c.Do(ctx, req, payload)
	return err
}

func (c *Client) Update(ctx context.Context, url string, id string, send interface{}, payload interface{}) error {
	url = fmt.Sprintf("%s/%s", url, id)
	req, err := c.NewRequest("PUT", url, send)
	if err != nil {
		return err
	}
	_, err = c.Do(ctx, req, payload)
	return err
}

// convert string id to ObjectId
func ToId(id string) bson.ObjectId {
	return bson.ObjectIdHex(id)
}

// convert ObjectId to string
func FromId(id bson.ObjectId) string {
	return id.Hex()
}

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.  API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.  Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		serviceError := &ServiceError{}
		if err := json.Unmarshal(data, serviceError); err == nil {
			errorResponse.ServiceError = serviceError
		}
	}
	return errorResponse
}

type ErrorResponse struct {
	Response     *http.Response
	ServiceError *ServiceError
}

func (r *ErrorResponse) Error() string {
	msg := "ErrorResponse: "
	if r.Response != nil {
		msg = fmt.Sprintf("%v %v: %d",
			r.Response.Request.Method,
			r.Response.Request.URL,
			r.Response.StatusCode)
	}
	if r.ServiceError != nil {
		msg = fmt.Sprintf("%s %s", msg, r.ServiceError)
	}
	return msg
}

// ServiceError is a transport object to pass information about a non-Http error occurred in a WebService while processing a request.
type ServiceError struct {
	Code    int
	Message string
}

// NewError returns a ServiceError using the code and reason
func NewError(code int, message string) ServiceError {
	return ServiceError{Code: code, Message: message}
}

// Error returns a text representation of the service error
func (s ServiceError) Error() string {
	return fmt.Sprintf("[ServiceError:%v] %v", s.Code, s.Message)
}

// parseBoolResponse determines the boolean result from a Bearded API response.
// Several Bearded API methods return boolean responses indicated by the HTTP
// status code in the response (true indicated by a 204, false indicated by a
// 404).  This helper function will determine that result and hide the 404
// error if present.  Any other error will be returned through as-is.
func parseBoolResponse(err error) (bool, error) {
	if err == nil {
		return true, nil
	}

	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode == http.StatusNotFound {
		// Simply false.  In this one case, we do not pass the error through.
		return false, nil
	}

	// some other real error occurred
	return false, err
}

// return true if http status code is 404 (Status not found)
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}

// return true if http status code is 409 (Status conflict)
func IsConflicted(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode == http.StatusConflict {
		return true
	}
	return false
}
