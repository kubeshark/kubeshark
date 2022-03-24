package oas

import (
	"bytes"
	"context"
	basenine "github.com/up9inc/basenine/client/go"
	"net"
	"testing"
	"time"
)

type fakeConn struct {
	sendBuffer    *bytes.Buffer
	receiveBuffer *bytes.Buffer
}

func (f fakeConn) Read(p []byte) (int, error)       { return f.sendBuffer.Read(p) }
func (f fakeConn) Write(p []byte) (int, error)      { return f.receiveBuffer.Write(p) }
func (fakeConn) Close() error                       { return nil }
func (fakeConn) LocalAddr() net.Addr                { return nil }
func (fakeConn) RemoteAddr() net.Addr               { return nil }
func (fakeConn) SetDeadline(t time.Time) error      { return nil }
func (fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (fakeConn) SetWriteDeadline(t time.Time) error { return nil }

func TestOASGen(t *testing.T) {
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

	_ = testCases

	gen := new(defaultOasGenerator)
	//fakeDB := new(FakeConn)
	gen.dbConn = new(basenine.Connection)
	gen.dbConn.Conn = &fakeConn{
		sendBuffer:    bytes.NewBufferString(""),
		receiveBuffer: bytes.NewBufferString(""),
	}
	ctx, cancel := context.WithCancel(context.Background())
	gen.ctx = ctx
	gen.cancel = cancel
	gen.runGenerator()
	/*
		for _, tc := range testCases {
			split := strings.Split(tc.inp, "/")
			pathObj := new(openapi.PathObj)
			node := tree.getOrSet(split, pathObj)

			fillPathParams(node, pathObj)

			if node.constant != nil && *node.constant != tc.label {
				t.Errorf("Constant does not match: %s != %s", *node.constant, tc.label)
			}

			if tc.numParams > 0 && (pathObj.Parameters == nil || len(*pathObj.Parameters) < tc.numParams) {
				t.Errorf("Wrong num of params, expected: %d", tc.numParams)
			}
		}*/
}
