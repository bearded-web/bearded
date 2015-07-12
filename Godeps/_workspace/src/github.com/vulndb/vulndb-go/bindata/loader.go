package bindata

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"sort"

	vulndb "github.com/vulndb/vulndb-go"
)

const (
	vulndbPath    = "vulndb/db"
	vulndbVersion = "vulndb/db-version.txt"
)

// Get all vulnerabilities from the bindata directory
// If you use this function then vulnerability data will be linked to your binary
func LoadFromBin() (vulndb.VulnList, error) {
	vulns := vulndb.VulnList{}
	assets, err := AssetDir(vulndbPath)
	if err != nil {
		return nil, err
	}
	for _, asset := range assets {
		assetPath := filepath.Join(vulndbPath, asset)
		assetInfo, err := AssetInfo(assetPath)
		if err != nil {
			return nil, err
		}
		if assetInfo.IsDir() {
			continue
		}
		vuln := &vulndb.Vuln{}
		data, err := Asset(assetPath)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, vuln); err != nil {
			return nil, err
		}
		vuln.Filename = asset
		vulns = append(vulns, vuln)
	}
	if err != nil {
		return nil, err
	}
	sort.Sort(vulns)
	return vulns, nil
}

func GetBinVersion() (string, error) {
	var version string
	data, err := Asset(filepath.Join(vulndbVersion))
	if err == nil {
		version = string(bytes.TrimSpace(data))
	}
	return version, err
}
