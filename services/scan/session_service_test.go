package scan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/services"
	"github.com/emicklei/go-restful"
	c "github.com/smartystreets/goconvey/convey"
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

func TestSessionCreate(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	testMgr := testMgr
	err := loadFixtures(testMgr)
	if err != nil {
		t.Fatal(err)
	}

	// create and auth user
	sess := filters.NewSession()
	u, err := testMgr.Users.Create(&user.User{})
	if err != nil {
		t.Fatal(err)
	}
	sess.Set(filters.SessionUserKey, u.Id.Hex())

	scanService := New(services.New(testMgr, nil, scheduler.NewFake(),
		email.NewConsoleBackend(), config.NewDispatcher().Api))
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	wsContainer.Filter(filters.SessionFilterMock(sess))
	scanService.Register(wsContainer)

	marshalSession := func(s *scan.Session) []byte {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	ts := httptest.NewServer(wsContainer)
	defer ts.Close()

	c.Convey("Given base scan with sessions", t, func() {
		projectObj, err := testMgr.Projects.Create(&project.Project{
			Name:  "default",
			Owner: u.Id,
		})
		c.So(err, c.ShouldBeNil)

		baseScan, err := testMgr.Scans.Create(&scan.Scan{
			Status:  scan.StatusWorking,
			Plan:    testMgr.NewId(),
			Target:  testMgr.NewId(),
			Owner:   u.Id,
			Project: projectObj.Id,
			Sessions: []*scan.Session{
				&scan.Session{
					Id:     testMgr.NewId(),
					Status: scan.StatusWorking,
					Plugin: testMgr.NewId(),
				},
			},
		})
		c.So(err, c.ShouldBeNil)
		scanId := testMgr.FromId(baseScan.Id)

		parentSession := baseScan.Sessions[0]

		baseScan.Sessions = append(baseScan.Sessions, parentSession)

		sess := &scan.Session{
			Step: &plan.WorkflowStep{
				Plugin: "barbudo/wappalyzer:0.0.2",
			},
			Scan:   baseScan.Id,
			Parent: parentSession.Id,
		}

		c.Convey("When create session with empty body", func() {
			res := sessionCreate(t, ts.URL, scanId, nil)
			shouldBeBadRequest(t, res, services.CodeWrongEntity, "wrong entity")
		})

		c.Convey("When create session with bad body", func() {
			res := sessionCreate(t, ts.URL, scanId, []byte("{'}"))
			shouldBeBadRequest(t, res, services.CodeWrongEntity, "wrong entity")
		})

		c.Convey("When create session without step", func() {
			res := sessionCreate(t, ts.URL, scanId, marshalSession(&scan.Session{}))
			shouldBeBadRequest(t, res, services.CodeWrongData, "step is required")
		})

		c.Convey("When create session with empty step.plugin", func() {
			sess.Step.Plugin = ""
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "step.plugin is required")
		})

		c.Convey("When create session with wrong scan", func() {
			sess.Scan = testMgr.NewId()
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "wrong scan id")
		})

		c.Convey("When scan is not in working state", func() {
			baseScan.Status = scan.StatusFinished
			if err := testMgr.Scans.Update(baseScan); err != nil {
				t.Fatal(err)
				return
			}
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "scan should have working status")
		})

		c.Convey("When parent is empty", func() {
			sess.Parent = ""
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "parent field is required")
		})

		c.Convey("When parent is not existed", func() {
			sess.Parent = testMgr.NewId()
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "parent not found")
		})

		c.Convey("When parent doesn't have working status", func() {
			parentSession.Status = scan.StatusFinished
			testMgr.Scans.UpdateSession(baseScan, parentSession)
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "parent should have working status")
		})

		c.Convey("When plugin has wrong name", func() {
			sess.Step.Plugin = "bad_name"
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))
			shouldBeBadRequest(t, res, services.CodeWrongData, "plugin bad_name is not found")
		})

		c.Convey("When everything is ok", func() {
			res := sessionCreate(t, ts.URL, scanId, marshalSession(sess))

			c.Convey("Session returns and updated into db", func() {
				c.So(res.StatusCode, c.ShouldEqual, http.StatusCreated)
				baseScan, err := testMgr.Scans.GetById(baseScan.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(baseScan.Sessions), c.ShouldEqual, 1)
				c.So(len(baseScan.Sessions[0].Children), c.ShouldEqual, 1)
				child := baseScan.Sessions[0].Children[0]
				c.So(child.Parent, c.ShouldEqual, parentSession.Id)
				c.So(child.Status, c.ShouldEqual, scan.StatusCreated)
				c.So(child.Scan, c.ShouldEqual, baseScan.Id)
			})
		})

	})
}

func shouldBeBadRequest(t *testing.T, res *http.Response, code services.CodeErr, message string) {
	c.Convey("Response should be 400 (Bad request)", func() {
		c.So(res.StatusCode, c.ShouldEqual, http.StatusBadRequest)
		e := getServiceError(t, res)
		c.So(e.Code, c.ShouldEqual, code)
		c.So(e.Message, c.ShouldEqual, message)
	})
}

func sessionCreate(t *testing.T, baseUrl string, scanId string, body []byte) *http.Response {
	url := fmt.Sprintf("%s/api/v1/scans/%s/sessions", baseUrl, scanId)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getServiceError(t *testing.T, res *http.Response) *restful.ServiceError {
	e := &restful.ServiceError{}
	err := json.NewDecoder(res.Body).Decode(e)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func loadFixtures(mgr *manager.Manager) error {
	{
		data := loadTestData("plugins.json")
		items := []*plugin.Plugin{}
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		for _, item := range items {
			item.Enabled = true
			if _, err := mgr.Plugins.Create(item); err != nil {
				return err
			}
		}
	}
	{
		data := loadTestData("plans.json")
		items := []*plan.Plan{}
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		for _, item := range items {
			if _, err := mgr.Plans.Create(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// test data
const testDataDir = "test_data"

func loadTestData(filename string) []byte {
	file := path.Join(testDataDir, filename)
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return raw
}
