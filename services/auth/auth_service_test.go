package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/emicklei/go-restful"
	c "github.com/smartystreets/goconvey/convey"

	"io/ioutil"
	"time"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/template"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/services"
	"github.com/stretchr/testify/require"
)

func TestResetPassword(t *testing.T) {
	mongo, dbName, err := tests.RandomTestMongoUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tests.RandomTestMongoDown(mongo, dbName)

	mgr := manager.New(mongo.DB(dbName))
	passCtx := passlib.NewContext()

	// create user
	u, err := mgr.Users.Create(&user.User{
		Email: "good@email.ru",
	})
	_ = u
	if err != nil {
		t.Fatal(err)
	}
	emailBackend := email.NewMemoryBackend(100)
	service := New(services.New(mgr, passCtx, scheduler.NewFake(),
		emailBackend, config.NewDispatcher().Api))
	service.Template = template.New(&template.Opts{Directory: "testdata/templates"})
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	service.Register(wsContainer)

	ts := httptest.NewServer(wsContainer)
	defer ts.Close()

	c.Convey("Given user", t, func() {

		c.Convey("Send empty email", func() {
			resp, err := resetPassword(ts.URL, &resetPasswordEntity{})
			c.Convey("Then result is bad request", func() {
				c.So(err, c.ShouldBeNil)
				c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
				sErr := getServiceError(t, resp)
				c.So(sErr.Code, c.ShouldEqual, services.CodeWrongData)
				c.So(sErr.Message, c.ShouldContainSubstring, "non zero value required")
			})
		})
		c.Convey("Send invalid email", func() {
			resp, err := resetPassword(ts.URL, &resetPasswordEntity{Email: "invalid"})
			c.Convey("Then result is bad request", func() {
				c.So(err, c.ShouldBeNil)
				c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
				sErr := getServiceError(t, resp)
				c.So(sErr.Code, c.ShouldEqual, services.CodeWrongData)
				c.So(sErr.Message, c.ShouldContainSubstring, "does not validate as email")

			})
		})
		c.Convey("Send non existed email", func() {
			resp, err := resetPassword(ts.URL, &resetPasswordEntity{Email: "bla@test.ru"})
			c.Convey("Then email isn't found", func() {
				c.So(err, c.ShouldBeNil)
				c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
				sErr := getServiceError(t, resp)
				c.So(sErr.Code, c.ShouldEqual, services.CodeWrongData)
				c.So(sErr.Message, c.ShouldContainSubstring, "not found")
			})
		})

		c.Convey("Send existed email", func() {
			resp, err := resetPassword(ts.URL, &resetPasswordEntity{Email: "good@email.ru"})
			c.Convey("Then response is 201", func() {
				c.So(err, c.ShouldBeNil)
				c.So(resp.StatusCode, c.ShouldEqual, http.StatusCreated)
				// wait for email
				var body []byte
				select {
				case msg := <-emailBackend.Messages():
					exportedMsg := msg.Export()
					body, err = ioutil.ReadAll(exportedMsg.Body)
					require.NoError(t, err)
				case <-time.After(time.Second * 1):
					t.Fatal("Timeout exceeded")
				}
				c.So(string(body), c.ShouldEqual, "Create new password")
			})
		})
	})

}

func resetPassword(baseUrl string, entity *resetPasswordEntity) (*http.Response, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/auth/reset-password", baseUrl))
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(entity)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", u.String(), buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	return resp, err
}

func getServiceError(t *testing.T, res *http.Response) *restful.ServiceError {
	e := &restful.ServiceError{}
	err := json.NewDecoder(res.Body).Decode(e)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
