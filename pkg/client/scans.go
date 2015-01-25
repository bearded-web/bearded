package client

import (
	"fmt"

	"code.google.com/p/go.net/context"

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
