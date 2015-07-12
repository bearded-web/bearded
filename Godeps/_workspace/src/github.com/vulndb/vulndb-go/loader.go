package vulndb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// get all vulnerabilities from the vulndb directory
func LoadFromDir(root string) (VulnList, error) {
	vulns := VulnList{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		vuln := &Vuln{}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, vuln); err != nil {
			return err
		}
		vuln.Filename = filepath.Base(path)
		vulns = append(vulns, vuln)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Sort(vulns)
	return vulns, nil
}
