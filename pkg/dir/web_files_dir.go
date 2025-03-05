package dir

import (
	"os"
	"strings"
)

func templates() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir, _, _ = strings.Cut(dir, "cmd")
	return dir + `templates\`
}

func Get_template_file(name string) string {
	return templates() + name
}
