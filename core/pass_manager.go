// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"container/list"
)

type Pass interface {
	RunOnBytes([]byte) ([]byte, error)
}

type PassManager struct {
	passes *list.List
}

func (this *PassManager) AddPass(p Pass) *PassManager {
	this.passes.PushBack(p)
	return this
}

func (this *PassManager) RunOnBytes(buf []byte) (result []byte, err error) {
	result = buf
	for e := this.passes.Front(); e != nil; e = e.Next() {
		p := e.Value.(Pass)
		result, err = p.RunOnBytes(result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func NewPassManager() *PassManager {
	return &PassManager{
		passes: list.New(),
	}
}

type PackUnpackPassManager struct {
	packPasses   *list.List
	unpackPasses *list.List
}

func (this *PackUnpackPassManager) AddPairedPasses(pack Pass, unpack Pass) {
	this.packPasses.PushBack(pack)
	this.packPasses.PushBack(unpack)
}

func (this *PackUnpackPassManager) CreatePackPassManager() *PassManager {
	pm := NewPassManager()
	for e := this.packPasses.Front(); e != nil; e = e.Next() {
		pm.AddPass(e.Value.(Pass))
	}
	return pm
}

func (this *PackUnpackPassManager) CreateUnpackPassManager() *PassManager {
	pm := NewPassManager()
	for e := this.unpackPasses.Back(); e != nil; e = e.Prev() {
		pm.AddPass(e.Value.(Pass))
	}
	return pm
}

func NewPackUnpackPassManager() *PackUnpackPassManager {
	return &PackUnpackPassManager{
		packPasses:   list.New(),
		unpackPasses: list.New(),
	}
}
