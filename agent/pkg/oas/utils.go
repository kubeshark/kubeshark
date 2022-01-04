package oas

import (
	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/shared/logger"
	"strings"
)

func initParams(obj **openapi.ParameterList) {
	if *obj == nil {
		var params openapi.ParameterList
		params = make([]openapi.Parameter, 0)
		*obj = &params
	}
}

func createSimpleParam(name string, in openapi.In, ptype openapi.SchemaType) *openapi.ParameterObj {
	required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
	schema := new(openapi.SchemaObj)
	schema.Type = make(openapi.Types, 0)
	schema.Type = append(schema.Type, ptype)
	newParam := openapi.ParameterObj{
		Name:     name,
		In:       in,
		Style:    "simple",
		Examples: map[string]openapi.Example{},
		Schema:   schema,
		Required: &required,
	}
	return &newParam
}

func findParamByName(params *openapi.ParameterList, in openapi.In, name string, caseInsensitive bool) (pathParam *openapi.ParameterObj) {
	for _, param := range *params {
		switch param.ParameterKind() {
		case openapi.ParameterKindReference:
			logger.Log.Warningf("Reference type is not supported for parameters")
		case openapi.ParameterKindObj:
			paramObj := param.(*openapi.ParameterObj)
			if paramObj.In != in {
				continue
			}

			if paramObj.Name == name || (caseInsensitive && strings.ToLower(paramObj.Name) == strings.ToLower(name)) {
				pathParam = paramObj
				break
			}
		}
	}
	return pathParam
}
