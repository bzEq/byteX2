// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"testing"
)

func TestPadd(t *testing.T) {
	pm := NewPassManager()
	pm.AddPass(&Padder{})
	pm.AddPass(&Unpadder{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

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
