package evaluate

import (
	"fmt"
	"k8strike/pkg/conf"
	"k8strike/pkg/util"
	"os"
	"path/filepath"
	"strings"
)

func SearchLocalFilePath() {

	filepath.Walk(conf.SensitiveFileConf.StartDir, func(path string, info os.FileInfo, err error) error {
		for _, name := range conf.SensitiveFileConf.NameList {
			currentPath := strings.ToLower(path)
			if strings.Contains(currentPath, name) {
				fmt.Printf("\t%s - %s\n", name, path)
				if util.IsDir(currentPath) {
					return filepath.SkipDir
				}
				return nil
			}
		}
		return nil
	})

}

func init() {
	RegisterSimpleCheck(CategorySensitiveFiles, "filesystem.sensitive", "Search for sensitive file paths", SearchLocalFilePath)
}
