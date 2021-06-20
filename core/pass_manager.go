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

func NewPassManagerWithPasses(passes *list.List) *PassManager {
	return &PassManager{
		passes: passes,
	}
}

type PackUnpackPassManager struct {
	packPasses   *list.List
	unpackPasses *list.List
}

func (this *PackUnpackPassManager) AddPairedPasses(pack Pass, unpack Pass) {
	this.packPasses.PushBack(pack)
	this.unpackPasses.PushFront(unpack)
}

func (this *PackUnpackPassManager) CreatePackPassManager() *PassManager {
	return NewPassManagerWithPasses(this.packPasses)
}

func (this *PackUnpackPassManager) CreateUnpackPassManager() *PassManager {
	return NewPassManagerWithPasses(this.unpackPasses)
}

func NewPackUnpackPassManager() *PackUnpackPassManager {
	return &PackUnpackPassManager{
		packPasses:   list.New(),
		unpackPasses: list.New(),
	}
}
