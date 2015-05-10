package target

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/emicklei/go-restful"
	c "github.com/smartystreets/goconvey/convey"

	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/services"
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
	})
}

func TestTargetService(t *testing.T) {
	testMgr := testMgr

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

	c.Convey("Given empty project", t, func() {
		c.So(err, c.ShouldBeNil)
		projectObj, err := testMgr.Projects.Create(&project.Project{
			Name:  "default",
			Owner: u.Id,
		})
		c.So(err, c.ShouldBeNil)

		c.Convey("Get list of all targets", func() {
			res, issues, err := getTargets(ts.URL, nil)
			c.Convey("Response should be empty", func() {
				c.So(err, c.ShouldBeNil)
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issues.Count, c.ShouldEqual, 0)
				c.So(len(issues.Results), c.ShouldEqual, 0)
			})
		})

		c.Convey("Create web target", func() {
			te := &TargetEntity{
				Type:    target.TypeWeb,
				Project: testMgr.FromId(projectObj.Id),
				Web: &WebTargetEntity{
					Domain: "http://google.com",
				},
			}
			c.Convey("With correct data", func() {
				res, tgt, err := createTarget(ts.URL, te)
				c.Convey("Successfull", func() {
					c.So(err, c.ShouldBeNil)
					c.So(res.StatusCode, c.ShouldEqual, http.StatusCreated)
					c.So(tgt.Type, c.ShouldEqual, target.TypeWeb)
					c.So(tgt.Project, c.ShouldEqual, projectObj.Id)
					c.So(tgt.Web.Domain, c.ShouldEqual, "http://google.com")
				})
			})
			c.Convey("With bad type", func() {
				te.Type = target.TargetType("bad type")
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Unknown target type")
			})
			c.Convey("Without web field", func() {
				te.Web = nil
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Web: zero value")
			})
			c.Convey("Without domain", func() {
				te.Web.Domain = ""
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Web.Domain: zero value")
			})
			c.Convey("With bad domain", func() {
				te.Web.Domain = "bad domain"
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "scheme must be http or https")
			})
			c.Convey("Without project id", func() {
				te.Project = ""
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Project: zero value")
			})
			c.Convey("With bad project id", func() {
				te.Project = "1234234sdf"
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Project: should be bson uuid in hex form")
			})
			c.Convey("With non existed project id", func() {
				te.Project = "553a8abcf18c0f18d3000004"
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Project doesn't exist")
			})
			c.Convey("With project without permission", func() {
				projectObj2, err := testMgr.Projects.Create(&project.Project{
					Name:  "default",
					Owner: projectObj.Id,
				})
				c.So(err, c.ShouldBeNil)
				te.Project = testMgr.FromId(projectObj2.Id)
				res, _, _ := createTarget(ts.URL, te)
				c.So(res.StatusCode, c.ShouldEqual, http.StatusForbidden)
			})
		})

		c.Convey("Create android target", func() {
			te := &TargetEntity{
				Type:    target.TypeAndroid,
				Project: testMgr.FromId(projectObj.Id),
				Android: &AndroidTargetEntity{
					Name: "First",
					File: &file.Meta{
						Id: "file id",
					},
				},
			}
			c.Convey("With correct data", func() {
				res, tgt, err := createTarget(ts.URL, te)
				c.Convey("Successfull", func() {
					c.So(err, c.ShouldBeNil)
					c.So(res.StatusCode, c.ShouldEqual, http.StatusCreated)
					c.So(tgt.Type, c.ShouldEqual, target.TypeAndroid)
					c.So(tgt.Project, c.ShouldEqual, projectObj.Id)
					c.So(tgt.Android.Name, c.ShouldEqual, "First")
					c.So(tgt.Android.File.Id, c.ShouldEqual, "file id")
					c.Convey("Update it with new file", func() {
						te.Android.File.Id = "file id2"
						te.Android.Name = "First2"
						res, tgt2, err := updateTarget(ts.URL, testMgr.FromId(tgt.Id), te)
						c.So(err, c.ShouldBeNil)
						c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
						c.So(tgt2.Type, c.ShouldEqual, target.TypeAndroid)
						c.So(tgt2.Project, c.ShouldEqual, projectObj.Id)
						c.So(tgt2.Android.Name, c.ShouldEqual, "First")
						c.So(tgt2.Android.File.Id, c.ShouldEqual, "file id2")

					})
				})
			})
			c.Convey("Without name", func() {
				te.Android.Name = ""
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Android.Name: zero value")
			})
			c.Convey("Without android field", func() {
				te.Android = nil
				res, _, _ := createTarget(ts.URL, te)
				shouldBeBadRequest(t, res, services.CodeWrongData, "Validation error: Android: zero value")
			})
			// TODO(m0sth8): add test for files permission violation
		})

	})

}

// Helpers

func getTargets(baseUrl string, val url.Values) (*http.Response, *target.TargetList, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/targets", baseUrl))
	if err != nil {
		return nil, nil, err
	}
	if val != nil || len(val) > 0 {
		u.RawQuery = val.Encode()
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	targets := &target.TargetList{}
	err = json.NewDecoder(resp.Body).Decode(targets)
	if err != nil {
		return nil, nil, err
	}
	return resp, targets, nil
}

func updateTarget(baseUrl string, id string, entity *TargetEntity) (*http.Response, *target.Target, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/targets/%s", baseUrl, id))
	if err != nil {
		return nil, nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(entity)
	if err != nil {
		return nil, nil, err
	}
	req, _ := http.NewRequest("PUT", u.String(), buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == http.StatusOK {
		obj := &target.Target{}
		err = json.NewDecoder(resp.Body).Decode(obj)
		if err != nil {
			return nil, nil, err
		}
		return resp, obj, nil
	}
	return resp, nil, nil
}

func createTarget(baseUrl string, entity *TargetEntity) (*http.Response, *target.Target, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/targets", baseUrl))
	if err != nil {
		return nil, nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(entity)
	if err != nil {
		return nil, nil, err
	}
	req, _ := http.NewRequest("POST", u.String(), buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == http.StatusCreated {
		obj := &target.Target{}
		err = json.NewDecoder(resp.Body).Decode(obj)
		if err != nil {
			return nil, nil, err
		}
		return resp, obj, nil
	}
	return resp, nil, nil
}

func shouldBeBadRequest(t *testing.T, res *http.Response, code services.CodeErr, message string) {
	c.Convey("Response should be 400 (Bad request)", func() {
		c.So(res.StatusCode, c.ShouldEqual, http.StatusBadRequest)
		e := getServiceError(t, res)
		c.So(e.Code, c.ShouldEqual, code)
		c.So(e.Message, c.ShouldEqual, message)
	})
}

func getServiceError(t *testing.T, res *http.Response) *restful.ServiceError {
	e := &restful.ServiceError{}
	err := json.NewDecoder(res.Body).Decode(e)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
