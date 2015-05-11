package issue

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
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/manager"
	"github.com/bearded-web/bearded/pkg/scheduler"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/bearded-web/bearded/pkg/utils"
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
	}())
}

func TestTargetIssues(t *testing.T) {
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

	c.Convey("Given empty issues collections", t, func() {
		_, err = testMgr.Issues.RemoveAll(bson.M{})
		c.So(err, c.ShouldBeNil)
		projectObj, err := testMgr.Projects.Create(&project.Project{
			Name:  "default",
			Owner: u.Id,
		})
		c.So(err, c.ShouldBeNil)
		targetObj, err := testMgr.Targets.Create(&target.Target{
			Project: projectObj.Id,
			Type:    target.TypeWeb,
		})
		c.So(err, c.ShouldBeNil)
		c.So(len(targetObj.SummaryReport.Issues), c.ShouldEqual, 0)

		c.Convey("Get list of all issues", func() {
			res, issues := getIssues(t, ts.URL, nil)
			c.Convey("Response should be empty", func() {
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issues.Count, c.ShouldEqual, 0)
				c.So(len(issues.Results), c.ShouldEqual, 0)
			})
		})

		c.Convey("Given target issue", func() {
			targetIssue, err := testMgr.Issues.Create(&issue.TargetIssue{
				Target:  targetObj.Id,
				Project: projectObj.Id,
				Issue: issue.Issue{
					Severity: issue.SeverityInfo,
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
			c.So(targetIssue.Severity, c.ShouldEqual, issue.SeverityInfo)
			err = testMgr.Targets.UpdateSummary(targetObj)
			c.So(err, c.ShouldBeNil)

			c.Convey("Get list of all issues", func() {
				res, issues := getIssues(t, ts.URL, nil)
				c.Convey("Response should have a new issue", func() {
					c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
					c.So(issues.Count, c.ShouldEqual, 1)
					c.So(len(issues.Results), c.ShouldEqual, 1)
					c.So(issues.Results[0].Id, c.ShouldEqual, targetIssue.Id)
				})
				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 1)
				c.So(targetObj2.SummaryReport.Issues[issue.SeverityInfo], c.ShouldEqual, 1)
			})

			c.Convey("Set issue status confirmed", func() {
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					StatusEntity: StatusEntity{
						Confirmed: utils.BoolP(false),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Confirmed, c.ShouldEqual, false)
				c.So(issueObj.Muted, c.ShouldEqual, false)
				c.So(issueObj.False, c.ShouldEqual, false)
				c.So(issueObj.Resolved, c.ShouldEqual, false)

			})

			c.Convey("Set issue status muted", func() {
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					StatusEntity: StatusEntity{
						Muted: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Confirmed, c.ShouldEqual, true)
				c.So(issueObj.Muted, c.ShouldEqual, true)
				c.So(issueObj.False, c.ShouldEqual, false)
				c.So(issueObj.Resolved, c.ShouldEqual, false)

				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 0)
			})

			c.Convey("Set issue status false", func() {
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					StatusEntity: StatusEntity{
						False: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Confirmed, c.ShouldEqual, true)
				c.So(issueObj.Muted, c.ShouldEqual, false)
				c.So(issueObj.False, c.ShouldEqual, true)
				c.So(issueObj.Resolved, c.ShouldEqual, false)

				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 0)
			})

			c.Convey("Set issue status resolved", func() {
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					StatusEntity: StatusEntity{
						Resolved: utils.BoolP(true),
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Confirmed, c.ShouldEqual, true)
				c.So(issueObj.Muted, c.ShouldEqual, false)
				c.So(issueObj.False, c.ShouldEqual, false)
				c.So(issueObj.Resolved, c.ShouldEqual, true)

				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 0)
			})

			c.Convey("Set issue severity to high", func() {
				high := issue.SeverityHigh
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					IssueEntity: IssueEntity{
						Severity: &high,
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Severity, c.ShouldEqual, issue.SeverityHigh)

				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 1)
				c.So(targetObj2.SummaryReport.Issues[issue.SeverityHigh], c.ShouldEqual, 1)
			})

			c.Convey("Set issue severity to bla", func() {
				bla := issue.Severity("bla")
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id), &TargetIssueEntity{
					IssueEntity: IssueEntity{
						Severity: &bla,
					},
				})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Severity, c.ShouldEqual, issue.SeverityInfo)
			})

			c.Convey("Set all issue status", func() {
				res, issueObj := updateIssue(t, ts.URL, testMgr.FromId(targetIssue.Id),
					&TargetIssueEntity{
						StatusEntity: StatusEntity{
							Muted:     utils.BoolP(true),
							Resolved:  utils.BoolP(true),
							False:     utils.BoolP(true),
							Confirmed: utils.BoolP(false),
						},
					})
				c.So(res.StatusCode, c.ShouldEqual, http.StatusOK)
				c.So(issueObj.Confirmed, c.ShouldEqual, false)
				c.So(issueObj.Muted, c.ShouldEqual, true)
				c.So(issueObj.False, c.ShouldEqual, true)
				c.So(issueObj.Resolved, c.ShouldEqual, true)
			})

			c.Convey("Create issue ", func() {
				res, issueObj, err := createIssue(t, ts.URL, &TargetIssueEntity{
					IssueEntity: IssueEntity{
						Summary: utils.StringP("First issue"),
						Vector: &VectorEntity{
							Url: "http://vector.url",
							HttpTransactions: []*HttpTransactionEntity{
								&HttpTransactionEntity{
									Id:     10,
									Url:    "http://vector.url?data=",
									Params: []string{"param1", "param2"},
									Method: "POST",
									Request: &HttpMyEntity{
										Status: "POST URL HTTP 1.0",
										Body: &issue.HttpBody{
											Content: "Content",
										},
										Header: []*HeaderMyEntity{
											&HeaderMyEntity{
												Key:    "key1",
												Values: []string{"val1", "val2"},
											},
										},
									},
									Response: &HttpMyEntity{
										Status: "200 OK",
										Body: &issue.HttpBody{
											Content: "response content",
										},
									},
								},
							},
						},
					},
					Target: testMgr.FromId(targetObj.Id),
				})
				if err != nil {
					t.Fatal(err)
				}
				c.So(res.StatusCode, c.ShouldEqual, http.StatusCreated)
				c.So(issueObj.UniqId, c.ShouldEqual, testMgr.FromId(issueObj.Id))
				c.So(issueObj.Summary, c.ShouldEqual, "First issue")
				c.So(issueObj.Severity, c.ShouldEqual, issue.SeverityInfo)
				c.So(issueObj.Target, c.ShouldEqual, targetObj.Id)
				c.So(issueObj.Project, c.ShouldEqual, projectObj.Id)
				c.So(issueObj.Confirmed, c.ShouldEqual, false)
				c.So(issueObj.Muted, c.ShouldEqual, false)
				c.So(issueObj.False, c.ShouldEqual, false)
				c.So(issueObj.Resolved, c.ShouldEqual, false)
				c.So(issueObj.Activities[0].Type, c.ShouldEqual, issue.ActivityReported)
				c.So(issueObj.Activities[0].User, c.ShouldEqual, u.Id)
				c.So(issueObj.Vector.Url, c.ShouldEqual, "http://vector.url")
				c.So(len(issueObj.Vector.HttpTransactions), c.ShouldEqual, 1)
				hTr := issueObj.Vector.HttpTransactions[0]
				c.So(hTr.Url, c.ShouldEqual, "http://vector.url?data=")
				c.So(hTr.Id, c.ShouldEqual, 10)
				c.So(hTr.Params[0], c.ShouldEqual, "param1")
				c.So(hTr.Method, c.ShouldEqual, "POST")
				c.So(hTr.Request.Status, c.ShouldEqual, "POST URL HTTP 1.0")
				c.So(hTr.Request.Body.Content, c.ShouldEqual, "Content")
				c.So(hTr.Request.Header.Get("key1"), c.ShouldEqual, "val1")
				c.So(hTr.Response.Status, c.ShouldEqual, "200 OK")
				c.So(hTr.Response.Body.Content, c.ShouldEqual, "response content")

				targetObj2, err := testMgr.Targets.GetById(targetObj.Id)
				c.So(err, c.ShouldBeNil)
				c.So(len(targetObj2.SummaryReport.Issues), c.ShouldEqual, 1)
				c.So(targetObj2.SummaryReport.Issues[issue.SeverityInfo], c.ShouldEqual, 2)
			})
			// TODO (m0sth8): test errors for creation
		})

	})

}

func TestIssuePermissions(t *testing.T) {
	// TODO (m0sth8): implement
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

func createIssue(t *testing.T, baseUrl string, entity *TargetIssueEntity) (*http.Response, *issue.TargetIssue, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/issues", baseUrl))
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
	defer resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == http.StatusCreated {
		issueObj := &issue.TargetIssue{}
		err = json.NewDecoder(resp.Body).Decode(issueObj)
		if err != nil {
			return nil, nil, err
		}
		return resp, issueObj, nil
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
