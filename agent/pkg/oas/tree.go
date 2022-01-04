package oas

import (
	"encoding/json"
	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/shared/logger"
	"strconv"
	"strings"
)

type NodePath = []string

type Node struct {
	constant *string
	param    *openapi.ParameterObj
	ops      *openapi.PathObj
	parent   *Node
	children []*Node
}

func (n *Node) getOrSet(path NodePath, pathObjToSet *openapi.PathObj) (node *Node) {
	if pathObjToSet == nil {
		panic("Invalid function call")
	}

	pathChunk := path[0]
	chunkIsParam := strings.HasPrefix(pathChunk, "{") && strings.HasSuffix(pathChunk, "}")
	chunkIsGibberish := isGibberish(pathChunk)

	var paramObj *openapi.ParameterObj
	if chunkIsParam && pathObjToSet != nil {
		paramObj = findParamByName(pathObjToSet.Parameters, openapi.InPath, pathChunk[1:len(pathChunk)-1], false)
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
			node.param = paramObj
		} else if chunkIsGibberish {
			initParams(&pathObjToSet.Parameters)

			newParam := n.createParam()
			node.param = newParam

			appended := append(*pathObjToSet.Parameters, newParam)
			pathObjToSet.Parameters = &appended
		} else {
			node.constant = &pathChunk
		}
	}

	// add example if it's a param
	if node.param != nil && !chunkIsParam {
		err := fillParamExample(node.param, pathChunk)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}
	}

	// TODO: eat up trailing slash, in a smart way: node.ops!=nil && path[1]==""
	if len(path) > 1 {
		return node.getOrSet(path[1:], pathObjToSet)
	} else if node.ops == nil {
		node.ops = pathObjToSet
	}

	return node
}

func fillParamExample(param *openapi.ParameterObj, exampleValue string) error {
	if param.Examples == nil {
		param.Examples = map[string]openapi.Example{}
	}

	cnt := 0
	for _, example := range param.Examples {
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

		if value == exampleValue || cnt > 5 { // 5 examples is enough
			return nil
		}
	}

	valMsg, err := json.Marshal(exampleValue)
	if err != nil {
		return err
	}

	param.Examples["example #"+strconv.Itoa(cnt)] = &openapi.ExampleObj{Value: valMsg}

	return nil
}

func (n *Node) createParam() *openapi.ParameterObj {
	newParam := createSimpleParam("param", "path", "string")
	x := n.countParentParams()
	if x > 1 {
		newParam.Name = newParam.Name + strconv.Itoa(x)
	}

	return newParam
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

func (n *Node) countParentParams() int {
	res := 0
	node := n
	for {
		if node.param != nil {
			res++
		}

		if node.parent == nil {
			break
		}
		node = node.parent
	}
	return res
}
