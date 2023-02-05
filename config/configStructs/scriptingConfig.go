package configStructs

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"github.com/robertkrimen/otto/parser"
)

type ScriptingConfig struct {
	Consts map[string]string `yaml:"consts"`
	Source string            `yaml:"source" default:""`
}

type Script struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

func (config *ScriptingConfig) GetScripts() (scripts []*Script, err error) {
	if config.Source == "" {
		return
	}

	var files []fs.FileInfo
	files, err = ioutil.ReadDir(config.Source)
	if err != nil {
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filename := f.Name()
		var body []byte
		body, err = os.ReadFile(filepath.Join(config.Source, filename))
		if err != nil {
			return
		}
		content := string(body)

		var program *ast.Program
		program, err = parser.ParseFile(nil, filename, content, parser.StoreComments)
		if err != nil {
			return
		}

		var title string
		code := content

		var idx0 file.Idx
		for node, comments := range program.Comments {
			if (idx0 > 0 && node.Idx0() > idx0) || len(comments) == 0 {
				continue
			}
			idx0 = node.Idx0()

			title = comments[0].Text
		}

		scripts = append(scripts, &Script{
			Title: title,
			Code:  code,
		})
	}

	return
}
