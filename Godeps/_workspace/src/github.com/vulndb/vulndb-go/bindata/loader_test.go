package bindata

import (
	"testing"
)

func TestLoadFromBin(t *testing.T) {
	vulnList, err := LoadFromBin()
	if err != nil {
		t.Fatalf("Expected no err but actual is %s", err)
	}
	if len(vulnList) == 0 {
		t.Fatalf("Expected vulnList length more then zero")
	}
	vuln := vulnList[0]
	if vuln.Id != 1 {
		t.Fatalf("Excepted first vuln with id 1 but actual is %d", vuln.Id)
	}
	if vuln.Title != "Allowed HTTP methods" {
		t.Fatalf("Excepted first vuln title `Allowed HTTP methods` but actual is `%s`", vuln.Title)
	}
}

func TestGetBinVersion(t *testing.T) {
	version, err := GetBinVersion()
	if err != nil {
		t.Fatalf("Expected no err but actual is %s", err)
	}
	// version must be in sha1 format
	if len(version) != 40 {
		t.Fatalf("Excepted version length is 40 but actual is %d", len(version))
	}
}
