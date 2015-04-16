package me

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	c "github.com/smartystreets/goconvey/convey"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/passlib"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/services"
)

func TestChangePassword(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	mongo, dbName, err := tests.RandomTestMongoUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tests.RandomTestMongoDown(mongo, dbName)

	mgr := manager.New(mongo.DB(dbName))

	passCtx := passlib.NewContext()

	// create and auth user
	sess := filters.NewSession()
	service := New(services.New(mgr, passCtx, scheduler.NewFake()))
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	wsContainer.Filter(filters.SessionFilterMock(sess))
	service.Register(wsContainer)

	ts := httptest.NewServer(wsContainer)
	defer ts.Close()

	c.Convey("Given authorized user with password - password", t, func() {
		pass, err := passCtx.Encrypt("password")
		if err != nil {
			t.Fatal(err)
		}
		u, err := mgr.Users.Create(&user.User{
			Password: pass,
		})
		if err != nil {
			t.Fatal(err)
		}
		sess.Set(filters.SessionUserKey, u.Id.Hex())

		c.Convey("Change password with wrong entity", func() {
			err, resp, sErr := changePassword(ts.URL, map[string]int{"old": 1})
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
			c.So(sErr, c.ShouldNotBeNil)
			c.So(sErr.Code, c.ShouldEqual, services.CodeWrongEntity)
		})

		c.Convey("Change password with wrong old password", func() {
			err, resp, sErr := changePassword(ts.URL, map[string]string{"old": "bad"})
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
			c.So(sErr, c.ShouldNotBeNil)
			c.So(sErr.Code, c.ShouldEqual, services.CodeWrongData)
		})

		c.Convey("Change password with right old password, but new is short", func() {
			err, resp, sErr := changePassword(ts.URL, map[string]string{"old": "password", "new": "short"})
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusBadRequest)
			c.So(sErr, c.ShouldNotBeNil)
			c.So(sErr.Code, c.ShouldEqual, services.CodeWrongData)
		})

		c.Convey("Change password with right old password, and good new password", func() {
			err, resp, sErr := changePassword(ts.URL, map[string]string{"old": "password", "new": "password2"})
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
			c.So(sErr, c.ShouldBeNil)

			modified, err := mgr.Users.GetById(u.Id)
			if err != nil {
				t.Fatal(err)
			}
			verified, err := passCtx.Verify("password2", modified.Password)
			if err != nil {
				t.Fatal(err)
			}
			c.So(verified, c.ShouldBeTrue)
		})
	})
}

func changePassword(baseUrl string, entity interface{}) (error, *http.Response, *restful.ServiceError) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/me/password", baseUrl))
	if err != nil {
		return err, nil, nil
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(entity)
	if err != nil {
		return err, nil, nil
	}
	req, _ := http.NewRequest("PUT", u.String(), buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err, nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		sErr := &restful.ServiceError{}
		err = json.NewDecoder(resp.Body).Decode(sErr)
		if err != nil {
			return err, nil, nil
		}
		return nil, resp, sErr
	}
	return nil, resp, nil

}
