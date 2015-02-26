package report

import "net/http"

type Url struct {
	Url string `json:"url"`
}

type Extra struct {
	Url   string `json:"url" description:""`
	Title string `json:"title"`
}

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
	Method   string      `json:"method"`
	Request  *HttpEntity `json:"request,omitempty"`
	Response *HttpEntity `json:"response,omitempty"`
}

type Issue struct {
	Severity         Severity           `json:"severity"`
	Summary          string             `json:"summary"`
	Desc             string             `json:"desc,omitempty"`
	Urls             []*Url             `json:"urls,omitempty" description:"where this issue is happened"`
	Extras           []*Extra           `json:"extras,omitempty" bson:"extras" description:"information about vulnerability"`
	HttpTransactions []*HttpTransaction `json:"httpTransactions,omitempty" bson:"httpTransactions"`
}
