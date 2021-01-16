// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"testing"
)

func TestCompress(t *testing.T) {
	pm := NewPassManager()
	pm.AddPass(&Compressor{})
	pm.AddPass(&Decompressor{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}
