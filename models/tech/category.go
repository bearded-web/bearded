//go:generate stringer -type=Category
package tech

type Category string

const (
	CMS                  = Category("cms")
	JavascriptFrameworks = Category("javascript-frameworks")
	Analytics            = Category("analytics")
	WebServers           = Category("web-servers")
	AdvertisingNetworks  = Category("advertising-networks")
	OperatingSystems     = Category("operating-systems")
)
