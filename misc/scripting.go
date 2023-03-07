package misc

import (
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"github.com/robertkrimen/otto/parser"
)

type Script struct {
	Path  string `json:"path"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

func ReadScriptFile(path string) (script *Script, err error) {
	filename := filepath.Base(path)
	var body []byte
	body, err = os.ReadFile(path)
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
	var titleIsSet bool
	code := content

	var idx0 file.Idx
	for node, comments := range program.Comments {
		if (titleIsSet && node.Idx0() > idx0) || len(comments) == 0 {
			continue
		}

		idx0 = node.Idx0()
		title = comments[0].Text
		titleIsSet = true
	}

	script = &Script{
		Path:  path,
		Title: title,
		Code:  code,
	}

	return
}
