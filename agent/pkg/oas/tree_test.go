package oas

import (
	"github.com/chanced/openapi"
	"strings"
	"testing"
)

func TestTree(t *testing.T) {
	testCases := []struct {
		inp string
	}{
		{"/"},
		{"/v1.0.0/config/launcher/sp_nKNHCzsN/f34efcae-6583-11eb-908a-00b0fcb9d4f6/vendor,init,conversation"},
	}

	tree := new(Node)
	for _, tc := range testCases {
		split := strings.Split(tc.inp, "/")
		node := tree.getOrSet(split, new(openapi.PathObj))

		if node.constant == nil {
			t.Errorf("nil constant: %s", tc.inp)
		}
	}
}
