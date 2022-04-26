package oas

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chanced/openapi"
)

func TestTree(t *testing.T) {
	testCases := []struct {
		inp       string
		numParams int
		label     string
	}{
		{"/", 0, ""},
		{"/v1.0.0/config/launcher/sp_nKNHCzsN/f34efcae-6583-11eb-908a-00b0fcb9d4f6/vendor,init,conversation", 1, "vendor,init,conversation"},
		{"/v1.0.0/config/launcher/sp_nKNHCzsN/{f34efcae-6583-11eb-908a-00b0fcb9d4f6}/vendor,init,conversation", 0, "vendor,init,conversation"},
		{"/getSvgs/size/small/brand/SFLY/layoutId/170943/layoutVersion/1/sizeId/742/surface/0/isLandscape/true/childSkus/%7B%7D", 1, "{}"},
	}

	tree := new(Node)
	for i, tc := range testCases {
		split := strings.Split(tc.inp, "/")
		pathObj := new(openapi.PathObj)
		node := tree.getOrSet(split, pathObj, fmt.Sprintf("%024d", i))

		fillPathParams(node, pathObj)

		if node.constant != nil && *node.constant != tc.label {
			t.Errorf("Constant does not match: %s != %s", *node.constant, tc.label)
		}

		if tc.numParams > 0 && (pathObj.Parameters == nil || len(*pathObj.Parameters) < tc.numParams) {
			t.Errorf("Wrong num of params, expected: %d", tc.numParams)
		}
	}
}
