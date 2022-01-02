package oas

import (
	"github.com/chanced/openapi"
)

type NodePath = []string

type Node struct {
	constant *string
	param    *openapi.ParameterObj
	ops      *openapi.PathObj
	children []*Node
}

func (n *Node) getOrSet(path NodePath, createIfMissing bool, pathObjToSet *openapi.PathObj) (node *Node, wasFound bool, wasSet bool) {
	// first pass is on constants
	for _, subnode := range n.children {
		if subnode.constant == nil {
			continue
		}

		if *subnode.constant == path[0] {
			node = subnode
		}
	}

	// first pass is on params
	for _, subnode := range n.children {
		if subnode.constant != nil {
			continue
		}

		// TODO: check the regex pattern?

		if node == nil {
			node = subnode
		}
	}

	wasFound = node != nil
	if !wasFound && createIfMissing {
		wasSet = true
		node = newNode()
		node.constant = &path[0]
		node.ops = pathObjToSet
		n.children = append(n.children, node)
	}

	// TODO: eat up trailing slash, in a smart way: node.ops!=nil && path[1]==""
	if len(path) > 1 {
		node, wasFound, wasSet = node.getOrSet(path[1:], createIfMissing, pathObjToSet)
	}

	return
}

func (n *Node) compact() {
	// TODO
}

func (n *Node) ListPaths() *openapi.Paths {
	paths := &openapi.Paths{Items: map[openapi.PathValue]*openapi.PathObj{}}

	// add self
	if n.ops != nil {
		paths.Items[openapi.PathValue("/"+*n.constant)] = n.ops
	}

	// recurse into children
	for _, child := range n.children {
		subPaths := child.ListPaths()
		for path, pathObj := range subPaths.Items {
			concat := *n.constant + string(path)
			paths.Items[openapi.PathValue(concat)] = pathObj
		}
	}

	return paths
}

func newNode() *Node {
	required := true // FFS! https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct/32364093
	param := openapi.ParameterObj{
		In:       "path",
		Style:    "simple",
		Required: &required,
		Examples: map[string]openapi.Example{},
	}
	return &Node{param: &param}
}
