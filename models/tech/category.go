package tech

import (
	"encoding/json"

	"github.com/bearded-web/bearded/pkg/pagination"
)

// TODO (m0sth8): generate from the file for all sdk

type Category string

const (
	CMS                       = Category("cms")
	JavascriptFrameworks      = Category("javascript-frameworks")
	Analytics                 = Category("analytics")
	WebServers                = Category("web-servers")
	AdvertisingNetworks       = Category("advertising-networks")
	OperatingSystems          = Category("operating-systems")
	MessageBoards             = Category("message-boards")
	DatabaseManagers          = Category("database-managers")
	DocumentationTools        = Category("documentation-tools")
	Widgets                   = Category("widgets")
	ECommerce                 = Category("ecommerce")
	PhotoGalleries            = Category("photo-galleries")
	Wikis                     = Category("wikis")
	HostingPanels             = Category("hosting-panels")
	Blogs                     = Category("blogs")
	IssueTrackers             = Category("issue-trackers")
	VideoPlayers              = Category("video-players")
	CommentSystems            = Category("comment-systems")
	Captchas                  = Category("captchas")
	FontScripts               = Category("font-scripts")
	WebFrameworks             = Category("web-frameworks")
	Miscellaneous             = Category("miscellaneous")
	Editors                   = Category("editors")
	LMS                       = Category("lms")
	CacheTools                = Category("cache-tools")
	RichTextEditors           = Category("rich-text-editors")
	JavascriptGraphics        = Category("javascript-graphics")
	MobileFrameworks          = Category("mobile-frameworks")
	ProgrammingLanguages      = Category("programming-languages")
	SearchEngines             = Category("search-engines")
	WebMail                   = Category("web-mail")
	CDN                       = Category("cdn")
	MarketingAutomation       = Category("marketing-automation")
	WebServerExtensions       = Category("web-server-extensions")
	Databases                 = Category("databases")
	Maps                      = Category("maps")
	NetworkDevices            = Category("network-devices")
	MediaServers              = Category("media-servers")
	WebCams                   = Category("webcams")
	Printers                  = Category("printers")
	PaymentProcessors         = Category("payment-processors")
	TagManagers               = Category("tag-managers")
	PayWalls                  = Category("paywalls")
	BuildCISystems            = Category("build-ci-systems")
	ControlSystems            = Category("control-systems")
	RemoteAccess              = Category("remote-access")
	DevTools                  = Category("dev-tools")
	NetworkStorage            = Category("network-storage")
	FeedReaders               = Category("feed-readers")
	DocumentManagementSystems = Category("document-management-systems")
)

var Categories = []Category{
	CMS,
	JavascriptFrameworks,
	Analytics,
	WebServers,
	AdvertisingNetworks,
	OperatingSystems,
	MessageBoards,
	DatabaseManagers,
	DocumentationTools,
	Widgets,
	ECommerce,
	PhotoGalleries,
	Wikis,
	HostingPanels,
	Blogs,
	IssueTrackers,
	VideoPlayers,
	CommentSystems,
	Captchas,
	FontScripts,
	WebFrameworks,
	Miscellaneous,
	Editors,
	LMS,
	CacheTools,
	RichTextEditors,
	JavascriptGraphics,
	MobileFrameworks,
	ProgrammingLanguages,
	SearchEngines,
	WebMail,
	CDN,
	MarketingAutomation,
	WebServerExtensions,
	Databases,
	Maps,
	NetworkDevices,
	MediaServers,
	WebCams,
	Printers,
	PaymentProcessors,
	TagManagers,
	PayWalls,
	BuildCISystems,
	ControlSystems,
	RemoteAccess,
	DevTools,
	NetworkStorage,
	FeedReaders,
	DocumentManagementSystems,
}

// It's a hack to show custom type as string in swagger
func (t Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t Category) Convert(text string) (interface{}, error) {
	return Category(text), nil
}

type CategoryList struct {
	pagination.Meta `json:",inline"`
	Results         []Category `json:"results"`
}
