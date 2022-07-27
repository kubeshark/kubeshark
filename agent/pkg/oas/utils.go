package oas

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/agent/pkg/har"

	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/logger"
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

func headerResolver(ref string) (*openapi.HeaderObj, error) {
	return nil, errors.New("JSON references are not supported at the moment: " + ref)
}

func initParams(obj **openapi.ParameterList) {
	if *obj == nil {
		var params openapi.ParameterList = make([]openapi.Parameter, 0)
		*obj = &params
	}
}

func initHeaders(respObj *openapi.ResponseObj) {
	if respObj.Headers == nil {
		var created openapi.Headers = map[string]openapi.Header{}
		respObj.Headers = created
	}
}

func createSimpleParam(name string, in openapi.In, ptype openapi.SchemaType) *openapi.ParameterObj {
	if name == "" {
		panic("Cannot create parameter with empty name")
	}
	required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
	schema := new(openapi.SchemaObj)
	schema.Type = openapi.Types{ptype}

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

func findParamByName(params *openapi.ParameterList, in openapi.In, name string) (idx int, pathParam *openapi.ParameterObj) {
	caseInsensitive := in == openapi.InHeader
	for i, param := range *params {
		idx = i
		paramObj, err := param.ResolveParameter(paramResolver)
		if err != nil {
			logger.Log.Warningf("Failed to resolve reference: %s", err)
			continue
		}

		if paramObj.In != in {
			continue
		}

		if paramObj.Name == name || (caseInsensitive && strings.EqualFold(paramObj.Name, name)) {
			pathParam = paramObj
			break
		}
	}

	return idx, pathParam
}

func findHeaderByName(headers *openapi.Headers, name string) *openapi.HeaderObj {
	for hname, param := range *headers {
		hdrObj, err := param.ResolveHeader(headerResolver)
		if err != nil {
			logger.Log.Warningf("Failed to resolve reference: %s", err)
			continue
		}

		if strings.EqualFold(hname, name) {
			return hdrObj
		}
	}
	return nil
}

type nvParams struct {
	In             openapi.In
	Pairs          []har.NVP
	IsIgnored      func(name string) bool
	GeneralizeName func(name string) string
}

func handleNameVals(gw nvParams, params **openapi.ParameterList, checkIgnore bool, sampleId string) {
	visited := map[string]*openapi.ParameterObj{}
	for _, pair := range gw.Pairs {
		if (checkIgnore && gw.IsIgnored(pair.Name)) || pair.Name == "" {
			continue
		}

		nameGeneral := gw.GeneralizeName(pair.Name)

		initParams(params)
		_, param := findParamByName(*params, gw.In, pair.Name)
		if param == nil {
			param = createSimpleParam(nameGeneral, gw.In, openapi.TypeString)
			appended := append(**params, param)
			*params = &appended
		}
		exmp := &param.Examples
		err := fillParamExample(&exmp, pair.Value)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}
		visited[nameGeneral] = param

		setSampleID(&param.Extensions, sampleId)
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

func createHeader(ptype openapi.SchemaType) *openapi.HeaderObj {
	required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
	schema := new(openapi.SchemaObj)
	schema.Type = make(openapi.Types, 0)
	schema.Type = append(schema.Type, ptype)

	style := openapi.StyleSimple
	newParam := openapi.HeaderObj{
		Style:    string(style),
		Examples: map[string]openapi.Example{},
		Schema:   schema,
		Required: &required,
	}
	return &newParam
}

func fillParamExample(param **openapi.Examples, exampleValue string) error {
	if **param == nil {
		**param = map[string]openapi.Example{}
	}

	cnt := 0
	for _, example := range **param {
		cnt++
		exampleObj, err := example.ResolveExample(exampleResolver)
		if err != nil {
			continue
		}

		var value string
		err = json.Unmarshal(exampleObj.Value, &value)
		if err != nil {
			logger.Log.Warningf("Failed decoding parameter example into string: %s", err)
			continue
		}

		if value == exampleValue || cnt >= 5 { // 5 examples is enough
			return nil
		}
	}

	valMsg, err := json.Marshal(exampleValue)
	if err != nil {
		return err
	}

	themap := **param
	themap["example #"+strconv.Itoa(cnt)] = &openapi.ExampleObj{Value: valMsg}

	return nil
}

// TODO: somehow generalize the two example setting functions, plus add body example handling

func addSchemaExample(existing *openapi.SchemaObj, bodyStr string) {
	if len(existing.Examples) < 5 {
		found := false
		for _, eVal := range existing.Examples {
			existingExample := ""
			err := json.Unmarshal(eVal, &existingExample)
			if err != nil {
				logger.Log.Debugf("Failed to unmarshal example: %v", eVal)
				continue
			}

			if existingExample == bodyStr {
				found = true
				break
			}
		}

		if !found {
			example, err := json.Marshal(bodyStr)
			if err != nil {
				logger.Log.Debugf("Failed to marshal example: %v", bodyStr)
				return
			}
			existing.Examples = append(existing.Examples, example)
		}
	}
}

func longestCommonXfix(strs [][]string, pre bool) []string { // https://github.com/jpillora/longestcommon
	empty := make([]string, 0)
	//short-circuit empty list
	if len(strs) == 0 {
		return empty
	}
	xfix := strs[0]
	//short-circuit single-element list
	if len(strs) == 1 {
		return xfix
	}
	//compare first to rest
	for _, str := range strs[1:] {
		xfixl := len(xfix)
		strl := len(str)
		//short-circuit empty strings
		if xfixl == 0 || strl == 0 {
			return empty
		}
		//maximum possible length
		maxl := xfixl
		if strl < maxl {
			maxl = strl
		}
		//compare letters
		if pre {
			//prefix, iterate left to right
			for i := 0; i < maxl; i++ {
				if xfix[i] != str[i] {
					xfix = xfix[:i]
					break
				}
			}
		} else {
			//suffix, iternate right to left
			for i := 0; i < maxl; i++ {
				xi := xfixl - i - 1
				si := strl - i - 1
				if xfix[xi] != str[si] {
					xfix = xfix[xi+1:]
					break
				}
			}
		}
	}
	return xfix
}

func longestCommonXfixStr(strs []string, pre bool) string { // https://github.com/jpillora/longestcommon
	//short-circuit empty list
	if len(strs) == 0 {
		return ""
	}
	xfix := strs[0]
	//short-circuit single-element list
	if len(strs) == 1 {
		return xfix
	}
	//compare first to rest
	for _, str := range strs[1:] {
		xfixl := len(xfix)
		strl := len(str)
		//short-circuit empty strings
		if xfixl == 0 || strl == 0 {
			return ""
		}
		//maximum possible length
		maxl := xfixl
		if strl < maxl {
			maxl = strl
		}
		//compare letters
		if pre {
			//prefix, iterate left to right
			for i := 0; i < maxl; i++ {
				if xfix[i] != str[i] {
					xfix = xfix[:i]
					break
				}
			}
		} else {
			//suffix, iternate right to left
			for i := 0; i < maxl; i++ {
				xi := xfixl - i - 1
				si := strl - i - 1
				if xfix[xi] != str[si] {
					xfix = xfix[xi+1:]
					break
				}
			}
		}
	}
	return xfix
}

func getSimilarPrefix(strs []string) string {
	chunked := make([][]string, 0)
	for _, item := range strs {
		chunked = append(chunked, strings.Split(item, "/"))
	}

	cmn := longestCommonXfix(chunked, true)
	res := make([]string, 0)
	for _, chunk := range cmn {
		if chunk != "api" && !IsVersionString(chunk) && !strings.HasPrefix(chunk, "{") {
			res = append(res, chunk)
		}
	}
	return strings.Join(res[1:], ".")
}

// returns all non-nil ops in PathObj
func getOps(pathObj *openapi.PathObj) []*openapi.Operation {
	ops := []**openapi.Operation{&pathObj.Get, &pathObj.Patch, &pathObj.Put, &pathObj.Options, &pathObj.Post, &pathObj.Trace, &pathObj.Head, &pathObj.Delete}
	res := make([]*openapi.Operation, 0)
	for _, opp := range ops {
		if *opp == nil {
			continue
		}
		res = append(res, *opp)
	}
	return res
}

// parses JSON into any possible value
func anyJSON(text string) (anyVal interface{}, isJSON bool) {
	isJSON = true
	asMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(text), &asMap)
	if err == nil && asMap != nil {
		return asMap, isJSON
	}

	asArray := make([]interface{}, 0)
	err = json.Unmarshal([]byte(text), &asArray)
	if err == nil && asArray != nil {
		return asArray, isJSON
	}

	asString := ""
	sPtr := &asString
	err = json.Unmarshal([]byte(text), &sPtr)
	if err == nil && sPtr != nil {
		return asString, isJSON
	}

	asInt := 0
	intPtr := &asInt
	err = json.Unmarshal([]byte(text), &intPtr)
	if err == nil && intPtr != nil {
		return asInt, isJSON
	}

	asFloat := 0.0
	floatPtr := &asFloat
	err = json.Unmarshal([]byte(text), &floatPtr)
	if err == nil && floatPtr != nil {
		return asFloat, isJSON
	}

	asBool := false
	boolPtr := &asBool
	err = json.Unmarshal([]byte(text), &boolPtr)
	if err == nil && boolPtr != nil {
		return asBool, isJSON
	}

	if text == "null" {
		return nil, isJSON
	}

	return nil, false
}

func cleanStr(str string, criterion func(r rune) bool) string {
	s := []byte(str)
	j := 0
	for _, b := range s {
		if criterion(rune(b)) {
			s[j] = b
			j++
		}
	}
	return string(s[:j])
}

/*
func isAlpha(s string) bool {
	for _, r := range s {
		if isAlphaRune(r) {
			return false
		}
	}
	return true
}
*/

func isAlphaRune(r rune) bool {
	return !((r < 'a' || r > 'z') && (r < 'A' || r > 'Z'))
}

func isAlNumRune(b rune) bool {
	return isAlphaRune(b) || ('0' <= b && b <= '9')
}

func deleteFromSlice(s []string, val string) []string {
	temp := s[:0]
	for _, x := range s {
		if x != val {
			temp = append(temp, x)
		}
	}
	return temp
}

func sliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func intersectSliceWithMap(required []string, names map[string]struct{}) []string {
	for name := range names {
		if !sliceContains(required, name) {
			required = deleteFromSlice(required, name)
		}
	}
	return required
}

func setSampleID(extensions *openapi.Extensions, id string) {
	if id != "" {
		if *extensions == nil {
			*extensions = openapi.Extensions{}
		}
		err := (extensions).SetExtension(SampleId, id)
		if err != nil {
			logger.Log.Warningf("Failed to set sample ID: %s", err)
		}
	}
}
