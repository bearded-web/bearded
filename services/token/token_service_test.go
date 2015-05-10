package token

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bearded-web/bearded/models/token"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/services"
	"github.com/emicklei/go-restful"
	c "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

var (
	testMgr *manager.Manager
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		mongo, dbName, err := tests.RandomTestMongoUp()
		if err != nil {
			println(err)
			os.Exit(1)
		}
		defer tests.RandomTestMongoDown(mongo, dbName)
		testMgr = manager.New(mongo.DB(dbName))
		return m.Run()
	}())
}

func TestTokens(t *testing.T) {
	// create and auth user
	sess := filters.NewSession()
	u, err := testMgr.Users.Create(&user.User{})
	if err != nil {
		t.Fatal(err)
	}
	sess.Set(filters.SessionUserKey, u.Id.Hex())

	service := New(services.New(testMgr, nil, scheduler.NewFake(),
		email.NewConsoleBackend(), config.NewDispatcher().Api))
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	wsContainer.Filter(filters.SessionFilterMock(sess))
	service.Register(wsContainer)

	ts := httptest.NewServer(wsContainer)
	defer ts.Close()

	api := client.NewClient(fmt.Sprintf("%s/api/", ts.URL), nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c.Convey("Given empty tokens collections", t, func() {
		_, err = testMgr.Tokens.RemoveAll(bson.M{})
		c.So(err, c.ShouldBeNil)

		c.Convey("Get list of all tokens", func() {
			tokens, err := api.Tokens.List(ctx, nil)
			c.So(err, c.ShouldBeNil)
			c.Convey("Response should be empty", func() {
				c.So(tokens.Count, c.ShouldEqual, 0)
				c.So(len(tokens.Results), c.ShouldEqual, 0)
			})
		})

		c.Convey("Create token", func() {
			t, err := api.Tokens.Create(ctx, &token.Token{Name: "name"})
			c.So(err, c.ShouldBeNil)
			c.So(t, c.ShouldNotBeNil)
			c.So(t.Hash, c.ShouldBeBlank)
			c.So(t.HashValue, c.ShouldNotBeBlank)

			c.Convey("Get list of all tokens", func() {
				tokens, err := api.Tokens.List(ctx, nil)
				c.So(err, c.ShouldBeNil)
				c.Convey("Response should has one value", func() {
					c.So(tokens.Count, c.ShouldEqual, 1)
					c.So(len(tokens.Results), c.ShouldEqual, 1)
				})
			})

			c.Convey("Get created token by id", func() {
				t2, err := api.Tokens.Get(ctx, client.FromId(t.Id))
				c.So(err, c.ShouldBeNil)
				c.Convey("Token shouldn't have hash value field", func() {
					c.So(t2.Id, c.ShouldEqual, t.Id)
					c.So(t2.Hash, c.ShouldBeBlank)
					c.So(t2.HashValue, c.ShouldBeBlank)
				})
			})

			c.Convey("When change token owner", func() {
				t.User = bson.NewObjectId()
				err := testMgr.Tokens.Update(t)
				c.So(err, c.ShouldBeNil)
				c.Convey("Get list of all tokens", func() {
					tokens, err := api.Tokens.List(ctx, nil)
					c.So(err, c.ShouldBeNil)
					c.Convey("Response should be empty", func() {
						c.So(tokens.Count, c.ShouldEqual, 0)
						c.So(len(tokens.Results), c.ShouldEqual, 0)
					})
				})
				c.Convey("Can't get token by id", func() {
					_, err := api.Tokens.Get(ctx, client.FromId(t.Id))
					c.So(client.IsNotFound(err), c.ShouldBeTrue)
				})

			})

			c.Convey("Remove token", func() {
				err := api.Tokens.Delete(ctx, client.FromId(t.Id))
				c.So(err, c.ShouldBeNil)

				c.Convey("Get list of all tokens", func() {
					tokens, err := api.Tokens.List(ctx, nil)
					c.So(err, c.ShouldBeNil)
					c.Convey("Response should be empty", func() {
						c.So(tokens.Count, c.ShouldEqual, 0)
						c.So(len(tokens.Results), c.ShouldEqual, 0)
					})
				})
				c.Convey("Token should have removed field", func() {
					t2, err := testMgr.Tokens.GetById(t.Id)
					c.So(err, c.ShouldBeNil)
					c.So(t2.Removed, c.ShouldBeTrue)
				})
				c.Convey("Can't get token by id", func() {
					_, err := api.Tokens.Get(ctx, client.FromId(t.Id))
					c.So(client.IsNotFound(err), c.ShouldBeTrue)
				})
			})
		})
	})
}
