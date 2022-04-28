package oas

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/logger"
)

type NodePath = []string

type Node struct {
	constant  *string
	pathParam *openapi.ParameterObj
	pathObj   *openapi.PathObj
	parent    *Node
	children  []*Node
}

func (n *Node) getOrSet(path NodePath, existingPathObj *openapi.PathObj, sampleId string) (node *Node) {
	if existingPathObj == nil {
		panic("Invalid function call")
	}

	pathChunk := path[0]
	potentialMatrix := strings.SplitN(pathChunk, ";", 2)
	if len(potentialMatrix) > 1 {
		pathChunk = potentialMatrix[0]
		logger.Log.Warningf("URI matrix params are not supported: %s", potentialMatrix[1])
	}

	chunkIsParam := strings.HasPrefix(pathChunk, "{") && strings.HasSuffix(pathChunk, "}")
	pathChunk, err := url.PathUnescape(pathChunk)
	if err != nil {
		logger.Log.Warningf("URI segment is not correctly encoded: %s", pathChunk)
		// any side effects on continuing?
	}

	chunkIsGibberish := IsGibberish(pathChunk) && !IsVersionString(pathChunk)

	var paramObj *openapi.ParameterObj
	if chunkIsParam && existingPathObj != nil && existingPathObj.Parameters != nil {
		_, paramObj = findParamByName(existingPathObj.Parameters, openapi.InPath, pathChunk[1:len(pathChunk)-1])
	}

	if paramObj == nil {
		node = n.searchInConstants(pathChunk)
	}

	if node == nil && pathChunk != "" {
		node = n.searchInParams(paramObj, pathChunk, chunkIsGibberish)
	}

	// still no node found, should create it
	if node == nil {
		node = new(Node)
		node.parent = n
		n.children = append(n.children, node)

		if paramObj != nil {
			node.pathParam = paramObj
		} else if chunkIsGibberish {
			newParam := n.createParam()
			node.pathParam = newParam
		} else {
			node.constant = &pathChunk
		}
	}

	if node.pathParam != nil {
		setSampleID(&node.pathParam.Extensions, sampleId)
	}

	// add example if it's a gibberish chunk
	if node.pathParam != nil && !chunkIsParam {
		exmp := &node.pathParam.Examples
		err := fillParamExample(&exmp, pathChunk)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}

		if len(*exmp) >= 3 && node.pathParam.Schema.Pattern == nil { // is it enough to decide on 2 samples?
			node.pathParam.Schema.Pattern = getPatternFromExamples(exmp)
		}
	}

	// TODO: eat up trailing slash, in a smart way: node.pathObj!=nil && path[1]==""
	if len(path) > 1 {
		return node.getOrSet(path[1:], existingPathObj, sampleId)
	} else if node.pathObj == nil {
		node.pathObj = existingPathObj
	}

	return node
}

func getPatternFromExamples(exmp *openapi.Examples) *openapi.Regexp {
	allInts := true
	strs := make([]string, 0)
	for _, example := range *exmp {
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
		strs = append(strs, value)

		if _, err := strconv.Atoi(value); err != nil {
			allInts = false
		}
	}

	if allInts {
		re := new(openapi.Regexp)
		re.Regexp = regexp.MustCompile(`\d+`)
		return re
	} else {
		prefix := longestCommonXfixStr(strs, true)
		suffix := longestCommonXfixStr(strs, false)

		pat := ""
		separators := "-._/:|*,+" // TODO: we could also cut prefix till the last separator
		if len(prefix) > 0 && strings.Contains(separators, string(prefix[len(prefix)-1])) {
			pat = "^" + regexp.QuoteMeta(prefix)
		}

		pat += ".+"

		if len(suffix) > 0 && strings.Contains(separators, string(suffix[0])) {
			pat += regexp.QuoteMeta(suffix) + "$"
		}

		if pat != ".+" {
			re := new(openapi.Regexp)
			re.Regexp = regexp.MustCompile(pat)
			return re
		}
	}
	return nil
}

func (n *Node) createParam() *openapi.ParameterObj {
	name := "param"

	if n.constant != nil { // the node is already a param
		// REST assumption, not always correct
		if strings.HasSuffix(*n.constant, "es") && len(*n.constant) > 4 {
			name = *n.constant
			name = name[:len(name)-2] + "Id"
		} else if strings.HasSuffix(*n.constant, "s") && len(*n.constant) > 3 {
			name = *n.constant
			name = name[:len(name)-1] + "Id"
		} else {
			name = *n.constant + "Id"
		}

		name = cleanStr(name, isAlNumRune)
		if !isAlphaRune(rune(name[0])) {
			name = "_" + name
		}
	}

	newParam := createSimpleParam(name, "path", "string")
	x := n.countParentParams()
	if x > 0 {
		newParam.Name = newParam.Name + strconv.Itoa(x)
	}

	return newParam
}

func (n *Node) searchInParams(paramObj *openapi.ParameterObj, chunk string, chunkIsGibberish bool) *Node {
	// look among params
	for _, subnode := range n.children {
		if subnode.constant != nil {
			continue
		}

		if paramObj != nil {
			// TODO: mergeParam(subnode.pathParam, paramObj)
			return subnode
		} else if subnode.pathParam.Schema.Pattern != nil { // it has defined param pattern, have to respect it
			// TODO: and not in exceptions
			if subnode.pathParam.Schema.Pattern.Match([]byte(chunk)) {
				return subnode
			} else if chunkIsGibberish {
				// TODO: what to do if gibberish chunk does not match the pattern and not in exceptions?
				return nil
			} else {
				return nil
			}
		} else if chunkIsGibberish {
			return subnode
		}

	}
	return nil
}

func (n *Node) searchInConstants(pathChunk string) *Node {
	// look among constants
	for _, subnode := range n.children {
		if subnode.constant == nil {
			continue
		}

		if *subnode.constant == pathChunk {
			return subnode
		}
	}
	return nil
}

func (n *Node) compact() {
	// TODO
}

func (n *Node) listPaths() *openapi.Paths {
	paths := &openapi.Paths{Items: map[openapi.PathValue]*openapi.PathObj{}}

	var strChunk string
	if n.constant != nil {
		strChunk = *n.constant
	} else if n.pathParam != nil {
		strChunk = "{" + n.pathParam.Name + "}"
	} // else -> this is the root node

	// add self
	if n.pathObj != nil {
		fillPathParams(n, n.pathObj)
		paths.Items[openapi.PathValue(strChunk)] = n.pathObj
	}

	// recurse into children
	for _, child := range n.children {
		subPaths := child.listPaths()
		for path, pathObj := range subPaths.Items {
			var concat string
			if n.parent == nil {
				concat = string(path)
			} else {
				concat = strChunk + "/" + string(path)
			}
			paths.Items[openapi.PathValue(concat)] = pathObj
		}
	}

	return paths
}

func fillPathParams(n *Node, pathObj *openapi.PathObj) {
	// collect all path parameters from parent hierarchy
	node := n
	for {
		if node.pathParam != nil {
			initParams(&pathObj.Parameters)

			idx, paramObj := findParamByName(pathObj.Parameters, openapi.InPath, node.pathParam.Name)
			if paramObj == nil {
				appended := append(*pathObj.Parameters, node.pathParam)
				pathObj.Parameters = &appended
			} else {
				(*pathObj.Parameters)[idx] = paramObj
			}
		}

		node = node.parent
		if node == nil {
			break
		}
	}
}

type PathAndOp struct {
	path string
	op   *openapi.Operation
}

func (n *Node) listOps() []PathAndOp {
	res := make([]PathAndOp, 0)
	for path, pathObj := range n.listPaths().Items {
		for _, op := range getOps(pathObj) {
			res = append(res, PathAndOp{path: string(path), op: op})
		}
	}
	return res
}

func (n *Node) countParentParams() int {
	res := 0
	node := n
	for {
		if node.pathParam != nil {
			res++
		}

		if node.parent == nil {
			break
		}
		node = node.parent
	}
	return res
}
