package oas

import (
	"errors"
	"github.com/chanced/openapi"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"strings"
)

func exampleResolver(ref string) (*openapi.ExampleObj, error) {
	return nil, errors.New("JSON references are not supported at the moment: " + ref)
}

func responseResolver(ref string) (*openapi.ResponseObj, error) {
	return nil, errors.New("JSON references are not supported at the moment: " + ref)
}

func reqBodyResolver(ref string) (*openapi.RequestBodyObj, error) {
	return nil, errors.New("JSON references are not supported at the moment: " + ref)
}

func paramResolver(ref string) (*openapi.ParameterObj, error) {
	return nil, errors.New("JSON references are not supported at the moment: " + ref)
}

func initParams(obj **openapi.ParameterList) {
	if *obj == nil {
		var params openapi.ParameterList
		params = make([]openapi.Parameter, 0)
		*obj = &params
	}
}

func createSimpleParam(name string, in openapi.In, ptype openapi.SchemaType) *openapi.ParameterObj {
	if name == "" {
		panic("aaa")
	}
	required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
	schema := new(openapi.SchemaObj)
	schema.Type = make(openapi.Types, 0)
	schema.Type = append(schema.Type, ptype)

	style := openapi.StyleSimple
	if in == openapi.InQuery {
		style = openapi.StyleForm
	}

	newParam := openapi.ParameterObj{
		Name:     name,
		In:       in,
		Style:    string(style),
		Examples: map[string]openapi.Example{},
		Schema:   schema,
		Required: &required,
	}
	return &newParam
}

func findParamByName(params *openapi.ParameterList, in openapi.In, name string) (pathParam *openapi.ParameterObj) {
	caseInsensitive := in == openapi.InHeader
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

type NVPair struct {
	Name  string
	Value string
}

type nvParams struct {
	In             openapi.In
	Pairs          func() []NVPair
	IsIgnored      func(name string) bool
	GeneralizeName func(name string) string
}

func qstrToNVP(list []har.QueryString) []NVPair {
	res := make([]NVPair, len(list))
	for idx, val := range list {
		res[idx] = NVPair{Name: val.Name, Value: val.Value}
	}
	return res
}

func hdrToNVP(list []har.Header) []NVPair {
	res := make([]NVPair, len(list))
	for idx, val := range list {
		res[idx] = NVPair{Name: val.Name, Value: val.Value}
	}
	return res
}

func handleNameVals(gw nvParams, params **openapi.ParameterList) {
	visited := map[string]*openapi.ParameterObj{}
	for _, pair := range gw.Pairs() {
		if gw.IsIgnored(pair.Name) {
			continue
		}

		nameGeneral := gw.GeneralizeName(pair.Name)

		initParams(params)
		param := findParamByName(*params, gw.In, pair.Name)
		if param == nil {
			param = createSimpleParam(nameGeneral, gw.In, openapi.TypeString)
			appended := append(**params, param)
			*params = &appended
		}
		err := fillParamExample(param, pair.Value)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}
		visited[nameGeneral] = param
	}

	// maintain "required" flag
	if *params != nil {
		for _, param := range **params {
			paramObj, err := param.ResolveParameter(paramResolver)
			if err != nil {
				logger.Log.Warningf("Failed to resolve param: %s", err)
				continue
			}
			if paramObj.In != gw.In {
				continue
			}

			_, ok := visited[strings.ToLower(paramObj.Name)]
			if !ok {
				flag := false
				paramObj.Required = &flag
			}
		}
	}
}
