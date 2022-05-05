package http

import (
	"encoding/json"

	"github.com/mertyildiran/gqlparser/v2/ast"
	"github.com/mertyildiran/gqlparser/v2/parser"
)

func isGraphQL(request map[string]interface{}) bool {
	if postData, ok := request["postData"].(map[string]interface{}); ok {
		if postData["mimeType"] == "application/json" {
			if text, ok := postData["text"].(string); ok {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(text), &data); err != nil {
					return false
				}

				if query, ok := data["query"].(string); ok {

					_, err := parser.ParseQuery(&ast.Source{Name: "ff", Input: query})
					if err == nil {
						return true
					}
				}
			}
		}
	}

	return false
}
