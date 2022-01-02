package oas

import (
	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/shared/logger"
	"strings"
)

type NodePath = []string

type Node struct {
	constant *string
	param    *openapi.ParameterObj
	ops      *openapi.PathObj
	children []*Node
}

func (n *Node) getOrSet(path NodePath, pathObjToSet *openapi.PathObj) (node *Node) {
	if pathObjToSet == nil {
		panic("Invalid function call")
	}

	pathChunk := path[0]
	chunkIsGibberish := isGibberish(pathChunk)

	paramObj := findPathParam(pathChunk, pathObjToSet)

	if paramObj == nil {
		node = n.searchInConstants(pathChunk)
	}

	if node == nil {
		node = n.searchInParams(paramObj, chunkIsGibberish)
	}

	// still no node found, should create it
	if node == nil {
		node = new(Node)
		n.children = append(n.children, node)

		if paramObj != nil {
			node.param = paramObj
		} else {
			required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
			newParam := openapi.ParameterObj{
				// the lack of Name keeps it invalid, until it's made valid below
				In:       "path",
				Style:    "simple",
				Required: &required,
				Examples: map[string]openapi.Example{},
			}

			if chunkIsGibberish {
				newParam.Name = "param"
			} else {
				node.constant = &pathChunk
			}

			node.param = &newParam
		}
	}

	// TODO: eat up trailing slash, in a smart way: node.ops!=nil && path[1]==""
	if len(path) > 1 {
		return node.getOrSet(path[1:], pathObjToSet)
	} else {
		node.ops = pathObjToSet
	}

	return node
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
				// TODO: mergeParam(subnode.param, paramObj)
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

func findPathParam(paramStrName string, pathObj *openapi.PathObj) (pathParam *openapi.ParameterObj) {
	if strings.HasPrefix(paramStrName, "{") && strings.HasSuffix(paramStrName, "}") && pathObj != nil {
		for _, param := range *pathObj.Parameters {
			switch param.ParameterKind() {
			case openapi.ParameterKindReference:
				logger.Log.Warningf("Reference type is not supported for parameters")
			case openapi.ParameterKindObj:
				paramObj := param.(*openapi.ParameterObj)
				if "{"+paramObj.Name+"}" == paramStrName {
					pathParam = paramObj
					break
				}
			}
		}
	}
	return pathParam
}

func (n *Node) compact() {
	// TODO
}

func (n *Node) listPaths() *openapi.Paths {
	paths := &openapi.Paths{Items: map[openapi.PathValue]*openapi.PathObj{}}

	var strChunk string
	if n.constant != nil {
		strChunk = *n.constant
	} else if n.param != nil {
		strChunk = "{" + n.param.Name + "}"
	} else {
		// this is the root node
	}

	// add self
	if n.ops != nil {
		paths.Items[openapi.PathValue(strChunk)] = n.ops
	}

	// recurse into children
	for _, child := range n.children {
		subPaths := child.listPaths()
		for path, pathObj := range subPaths.Items {
			var concat string
			if n.param == nil {
				concat = string(path)
			} else {
				concat = strChunk + "/" + string(path)
			}
			paths.Items[openapi.PathValue(concat)] = pathObj
		}
	}

	return paths
}
