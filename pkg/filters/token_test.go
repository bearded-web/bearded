package filters

import (
	"fmt"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/tests"
	c "github.com/smartystreets/goconvey/convey"
)

var (
	testMgr *manager.Manager
)

func TestMain(m *testing.M) {
	mongo, dbName, err := tests.RandomTestMongoUp()
	if err != nil {
		println(err)
		os.Exit(1)
	}
	defer tests.RandomTestMongoDown(mongo, dbName)
	testMgr = manager.New(mongo.DB(dbName))
	os.Exit(m.Run())
}

func TestGetUserByToken(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	u, err := testMgr.Users.Create(&user.User{})
	if err != nil {
		t.Fatal(err)
	}

	c.Convey("Given user and his token", t, func() {
		token, err := testMgr.Tokens.GetOrCreate(u.Id)
		c.So(err, c.ShouldBeNil)
		hash := token.Hash
		c.Convey("Take user by good hash", func() {
			u2 := getUserByToken(testMgr, fmt.Sprintf("Bearer %s", hash))
			c.So(u, c.ShouldNotBeNil)
			c.So(u.Id, c.ShouldEqual, u2.Id)
		})
		c.Convey("Take user by wrong hash", func() {
			u2 := getUserByToken(testMgr, fmt.Sprintf("Bearer 2%s", hash[1:]))
			c.So(u2, c.ShouldBeNil)
		})
		c.Convey("Take user by wrong first part", func() {
			u2 := getUserByToken(testMgr, fmt.Sprintf("Auth %s", hash))
			c.So(u2, c.ShouldBeNil)
		})
		c.Convey("Take user by empty auth", func() {
			u2 := getUserByToken(testMgr, fmt.Sprintf(""))
			c.So(u2, c.ShouldBeNil)
		})

	})

}
