package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestFilesDownload(t *testing.T) {
	bg := context.Background()
	handlerMock := func(res http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/v1/files/1/download":
			res.Header().Add("Content-Type", "octet/stream")
			res.Write([]byte("data 1"))
		case "/v1/files/2/download":
			http.NotFound(res, req)
		}
	}
	s := httptest.NewServer(http.HandlerFunc(handlerMock))
	defer s.Close()
	baseUrl, _ := url.Parse(s.URL)

	client := &Client{
		BaseURL: baseUrl,
		client:  http.DefaultClient,
		Debug:   true,
	}
	f := &FilesService{client: client}

	{
		buf, err := f.Download(utils.JustTimeout(bg, time.Millisecond*10), "1")
		require.NoError(t, err)
		require.NotNil(t, buf)
		data, err := ioutil.ReadAll(buf)
		require.NoError(t, err)
		require.Equal(t, "data 1", string(data))
	}
	{
		_, err := f.Download(utils.JustTimeout(bg, time.Millisecond*10), "2")
		require.Error(t, err)
	}

}
