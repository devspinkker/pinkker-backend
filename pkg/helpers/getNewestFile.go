package helpers

import (
	"os"
	"sort"
	"strings"
)

func GetNewestFile(files []os.FileInfo, path string) string {
	var out []map[string]interface{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp4") {
			out = append(out, map[string]interface{}{
				"file":  file.Name(),
				"mtime": file.ModTime().UnixNano(),
			})
		}
	}

	sortFilesByMTime(out)

	if len(out) > 0 {
		return out[0]["file"].(string)
	}

	return ""
}

func sortFilesByMTime(files []map[string]interface{}) {
	sort.Slice(files, func(i, j int) bool {
		return files[i]["mtime"].(int64) > files[j]["mtime"].(int64)
	})
}
