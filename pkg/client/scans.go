package client

import (
	"fmt"

	"code.google.com/p/go.net/context"

	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/models/scan"
)

const scansUrl = "scans"

type ScansService struct {
	client *Client
}

func (s *ScansService) String() string {
	return Stringify(s)
}

type ScansListOpts struct {
	Name string `url:"name"`
}

// List scans.
//
//
func (s *ScansService) List(ctx context.Context, opt *ScansListOpts) (*scan.ScanList, error) {
	scanList := &scan.ScanList{}
	return scanList, s.client.List(ctx, scansUrl, opt, scanList)
}

func (s *ScansService) Get(ctx context.Context, id string) (*scan.Scan, error) {
	scan := &scan.Scan{}
	return scan, s.client.Get(ctx, scansUrl, id, scan)
}

func (s *ScansService) Update(ctx context.Context, src *scan.Scan) (*scan.Scan, error) {
	pl := &scan.Scan{}
	id := FromId(src.Id)
	return pl, s.client.Update(ctx, scansUrl, id, src, pl)
}

func (s *ScansService) SessionUpdate(ctx context.Context, src *scan.Session) (*scan.Session, error) {
	obj := &scan.Session{}
	scanId := FromId(src.Scan)
	id := FromId(src.Id)
	sessUrl := fmt.Sprintf("%s/%s/sessions", scansUrl, scanId)
	return obj, s.client.Update(ctx, sessUrl, id, src, obj)
}

func (s *ScansService) SessionGet(ctx context.Context, scanId, sessionId string) (*scan.Session, error) {
	obj := &scan.Session{}
	sessUrl := fmt.Sprintf("%s/%s/sessions", scansUrl, scanId)
	return obj, s.client.Get(ctx, sessUrl, sessionId, obj)
}

func (s *ScansService) SessionAddChild(ctx context.Context, child *scan.Session) (*scan.Session, error) {
	scanId := FromId(child.Scan)
	obj := &scan.Session{}
	sessUrl := fmt.Sprintf("%s/%s/sessions", scansUrl, scanId)
	return obj, s.client.Create(ctx, sessUrl, child, obj)
}

func (s *ScansService) SessionReportCreate(ctx context.Context,
	src *scan.Session, rep *report.Report) (*report.Report, error) {

	obj := &report.Report{}
	scanId := FromId(src.Scan)
	id := FromId(src.Id)
	reportUrl := fmt.Sprintf("%s/%s/sessions/%s/report", scansUrl, scanId, id)
	return obj, s.client.Create(ctx, reportUrl, rep, obj)
}

func (s *ScansService) SessionReportGet(ctx context.Context, sc *scan.Session) (*report.Report, error) {
	obj := &report.Report{}
	scanId := FromId(sc.Scan)
	id := FromId(sc.Id)
	reportUrl := fmt.Sprintf("%s/%s/sessions/%s", scansUrl, scanId, id)
	println("get report url", reportUrl)
	return obj, s.client.Get(ctx, reportUrl, "report", obj)

}
