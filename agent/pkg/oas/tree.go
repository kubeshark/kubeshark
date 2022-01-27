package oas

import (
	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/shared/logger"
	"net/url"
	"strconv"
	"strings"
)

type NodePath = []string

type Node struct {
	constant  *string
	pathParam *openapi.ParameterObj
	pathObj   *openapi.PathObj
	parent    *Node
	children  []*Node
}

func (n *Node) getOrSet(path NodePath, existingPathObj *openapi.PathObj) (node *Node) {
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
		paramObj = findParamByName(existingPathObj.Parameters, openapi.InPath, pathChunk[1:len(pathChunk)-1])
	}

	if paramObj == nil {
		node = n.searchInConstants(pathChunk)
	}

	if node == nil {
		node = n.searchInParams(paramObj, chunkIsGibberish)
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

			//initParams(&existingPathObj.Parameters)
			//appended := append(*existingPathObj.Parameters, newParam)
			//existingPathObj.Parameters = &appended
		} else {
			node.constant = &pathChunk
		}
	}

	// add example if it's a gibberish chunk
	if node.pathParam != nil && !chunkIsParam {
		exmp := &node.pathParam.Examples
		err := fillParamExample(&exmp, pathChunk)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}
	}

	// TODO: eat up trailing slash, in a smart way: node.pathObj!=nil && path[1]==""
	if len(path) > 1 {
		return node.getOrSet(path[1:], existingPathObj)
	} else if node.pathObj == nil {
		node.pathObj = existingPathObj
	}

	return node
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
		} else if isAlpha(*n.constant) {
			name = *n.constant + "Id"
		}

		name = cleanNonAlnum([]byte(name))
	}

	newParam := createSimpleParam(name, "path", "string")
	x := n.countParentParams()
	if x > 0 {
		newParam.Name = newParam.Name + strconv.Itoa(x)
	}

	return newParam
}

func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func (n *Node) searchInParams(paramObj *openapi.ParameterObj, chunkIsGibberish bool) *Node {
	// look among params
	if paramObj != nil || chunkIsGibberish {
		for _, subnode := range n.children {
			if subnode.constant != nil {
				continue
			}

			// TODO: check the regex pattern of param? for exceptions etc

			if paramObj != nil {
				// TODO: mergeParam(subnode.pathParam, paramObj)
				return subnode
			} else {
				return subnode
			}
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
	} else {
		// this is the root node
	}

	// add self
	if n.pathObj != nil {
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
