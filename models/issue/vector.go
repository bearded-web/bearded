package issue

import "net/http"

type HttpBody struct {
	ContentEncoding string `json:"contentEncoding"`
	Content         []byte `json:"content"`
}

type HttpEntity struct {
	Status string      `json:"status"`
	Header http.Header `json:"header"`
	Body   *HttpBody   `json:"body,omitempty"`
}

type HttpTransaction struct {
	Id       int         `json:"id,omitempty"`
	Url      string      `json:"url"`
	Params   []string    `json:"params,omitempty"`
	Method   string      `json:"method"`
	Request  *HttpEntity `json:"request,omitempty"`
	Response *HttpEntity `json:"response,omitempty"`
}

type Vector struct {
	Url              string             `json:"url,omitempty" description:"where this issue is happened"`
	HttpTransactions []*HttpTransaction `json:"httpTransactions,omitempty" bson:"httpTransactions"`
}

type Vectors []*Vector
