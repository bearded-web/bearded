package issue

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
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/pkg/utils"
	"github.com/bearded-web/bearded/services"
)

func TestSessionCreate(t *testing.T) {
	mongo, dbName, err := tests.RandomTestMongoUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tests.RandomTestMongoDown(mongo, dbName)

	mgr := manager.New(mongo.DB(dbName))

	// create and auth user
	sess := filters.NewSession()
	u, err := mgr.Users.Create(&user.User{})
	if err != nil {
		t.Fatal(err)
	}
	sess.Set(filters.SessionUserKey, u.Id.Hex())

	service := New(services.New(mgr, nil, scheduler.NewFake()))
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	wsContainer.Filter(filters.SessionFilterMock(sess))
	service.Register(wsContainer)

	ts := httptest.NewServer(wsContainer)
	defer ts.Close()

	c.Convey("Given empty issues collections", t, func() {
		_, err = mongo.DB(dbName).C("issues").RemoveAll(bson.M{})
		c.So(err, c.ShouldBeNil)
		projectObj, err := mgr.Projects.Create(&project.Project{
			Name:  "default",
			Owner: u.Id,
		})
		c.So(err, c.ShouldBeNil)
		targetObj, err := mgr.Targets.Create(&target.Target{
			Project: projectObj.Id,
			Type:    target.TypeWeb,
		})
		c.So(err, c.ShouldBeNil)

		c.Convey("Get list of all issues", func() {
			res, issues := getIssues(t, ts.URL, nil)
			c.Convey("Response should be empty", func() {
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issues.Count, c.ShouldEqual, 0)
				c.So(len(issues.Results), c.ShouldEqual, 0)
			})
		})

		c.Convey("Given target issue", func() {
			targetIssue, err := mgr.Issues.Create(&issue.TargetIssue{
				Target:  targetObj.Id,
				Project: projectObj.Id,
				Issue: issue.Issue{
					UniqId: "1",
				},
				Status: issue.Status{
					Confirmed: true,
				},
			})
			c.So(err, c.ShouldBeNil)
			c.So(targetIssue.Confirmed, c.ShouldEqual, true)
			c.So(targetIssue.Muted, c.ShouldEqual, false)
			c.So(targetIssue.False, c.ShouldEqual, false)
			c.So(targetIssue.Resolved, c.ShouldEqual, false)

			c.Convey("Get list of all issues", func() {
				res, issues := getIssues(t, ts.URL, nil)
				c.Convey("Response should have a new issue", func() {
					c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
					c.So(issues.Count, c.ShouldEqual, 1)
					c.So(len(issues.Results), c.ShouldEqual, 1)
					c.So(issues.Results[0].Id, c.ShouldEqual, targetIssue.Id)
				})
			})

			c.Convey("Set issue status confirmed", func() {
				res, issue := updateIssue(t, ts.URL, mgr.FromId(targetIssue.Id), &TargetIssueEntity{
					Status: Status{
						Confirmed: utils.BoolP(false),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issue.Confirmed, c.ShouldEqual, false)
				c.So(issue.Muted, c.ShouldEqual, false)
				c.So(issue.False, c.ShouldEqual, false)
				c.So(issue.Resolved, c.ShouldEqual, false)
			})

			c.Convey("Set issue status muted", func() {
				res, issue := updateIssue(t, ts.URL, mgr.FromId(targetIssue.Id), &TargetIssueEntity{
					Status: Status{
						Muted: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issue.Confirmed, c.ShouldEqual, true)
				c.So(issue.Muted, c.ShouldEqual, true)
				c.So(issue.False, c.ShouldEqual, false)
				c.So(issue.Resolved, c.ShouldEqual, false)
			})

			c.Convey("Set issue status false", func() {
				res, issue := updateIssue(t, ts.URL, mgr.FromId(targetIssue.Id), &TargetIssueEntity{
					Status: Status{
						False: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issue.Confirmed, c.ShouldEqual, true)
				c.So(issue.Muted, c.ShouldEqual, false)
				c.So(issue.False, c.ShouldEqual, true)
				c.So(issue.Resolved, c.ShouldEqual, false)
			})

			c.Convey("Set issue status resolved", func() {
				res, issue := updateIssue(t, ts.URL, mgr.FromId(targetIssue.Id), &TargetIssueEntity{
					Status: Status{
						Resolved: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issue.Confirmed, c.ShouldEqual, true)
				c.So(issue.Muted, c.ShouldEqual, false)
				c.So(issue.False, c.ShouldEqual, false)
				c.So(issue.Resolved, c.ShouldEqual, true)
			})

			c.Convey("Set all issue status", func() {
				res, issue := updateIssue(t, ts.URL, mgr.FromId(targetIssue.Id), &TargetIssueEntity{
					Status: Status{
						Muted:     utils.BoolP(true),
						Resolved:  utils.BoolP(true),
						False:     utils.BoolP(true),
						Confirmed: utils.BoolP(false),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issue.Confirmed, c.ShouldEqual, false)
				c.So(issue.Muted, c.ShouldEqual, true)
				c.So(issue.False, c.ShouldEqual, true)
				c.So(issue.Resolved, c.ShouldEqual, true)
			})

		})

	})

}

// Helpers

func getIssues(t *testing.T, baseUrl string, val url.Values) (*http.Response, *issue.TargetIssueList) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/issues", baseUrl))
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	issues := &issue.TargetIssueList{}
	err = json.NewDecoder(resp.Body).Decode(issues)
	if err != nil {
		t.Fatal(err)
	}
	return resp, issues
}

func updateIssue(t *testing.T, baseUrl string, id string, entity *TargetIssueEntity) (*http.Response, *issue.TargetIssue) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/issues/%s", baseUrl, id))
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(entity)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("PUT", u.String(), buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusOK {
		issueObj := &issue.TargetIssue{}
		err = json.NewDecoder(resp.Body).Decode(issueObj)
		if err != nil {
			t.Fatal(err)
		}
		return resp, issueObj
	}
	return resp, nil

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
