// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bytes"
	"testing"
)

func TestLV(t *testing.T) {
	src := bytes.NewBuffer([]byte("wtf"))
	dst := &bytes.Buffer{}
	p := &LVPacker{}
	err := p.Translate(src, dst)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	src.Reset()
	up := &LVUnpacker{}
	err = up.Translate(dst, src)
	if string(src.Bytes()) != "wtf" || err != nil {
		t.Log(err)
		t.Log(src)
		t.Fail()
	}
}
