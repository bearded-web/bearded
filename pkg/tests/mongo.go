package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/bearded-web/bearded/pkg/utils"
	mgo "gopkg.in/mgo.v2"
)

const TestDbPrefix = "test_"

// Connect to mongo using addr, which must contain db name (mongodb://localhost/test_NAME)
// Be careful, database will be dropped.
func PrepareTestDb(addr string) (*mgo.Session, error) {
	s, err := mgo.DialWithTimeout(addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	// check if the default db has prefix "test_"
	if !strings.HasPrefix(s.DB("").Name, TestDbPrefix) {
		s.Close()
		return nil, fmt.Errorf("Default db name for test should be with prefix %s", TestDbPrefix)
	}
	// clean test db
	s.DB("").DropDatabase()
	return s, nil
}

// Create db with randomly generated name.
// Returns session and database name
func RandomTestMongoUp() (*mgo.Session, string, error) {
	dbName := fmt.Sprintf("%s%s", TestDbPrefix, utils.UuidV4String())
	session, err := PrepareTestDb(fmt.Sprintf("mongodb://localhost/%s", dbName))
	return session, dbName, err
}

// Remove testing database and close session
func RandomTestMongoDown(session *mgo.Session, dbName string) {
	session.DB(dbName).DropDatabase()
	session.Close()
}
