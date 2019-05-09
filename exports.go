package cnbshim

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func DumpExportsFile(exportsData, envDir string) error {
	lines := strings.SplitN(exportsData, "\n", 2)

	for _, line := range lines {
		if found, key, value := ParseExportsFileLine(line); found {
			err := WriteEnvFile(envDir, key, value[1 : len(value)-1])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ParseExportsFileLine(line string) (bool, string, string) {
	components := strings.Split(line, "=")
	export := strings.Split(components[0], " ")
	if strings.TrimSpace(export[0]) == "export" {
		return true, strings.TrimSpace(export[1]), components[1]
	} else {
		return false, "", ""
	}

}

func WriteEnvFile(envDir, filename, value string) error {
	return ioutil.WriteFile(filepath.Join(envDir, filename), []byte(value), 0644)
}