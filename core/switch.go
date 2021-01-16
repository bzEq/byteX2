// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"net"
)

// SimpleSwitch is not responsible to close red and blue.
type SimpleSwitch struct {
	doneRB, doneBR chan struct{}
	red, blue      net.Conn
	rbt, brt       Translator
}

func NewSimpleSwitch(red, blue net.Conn, rbt, brt Translator) *SimpleSwitch {
	return &SimpleSwitch{
		doneRB: make(chan struct{}),
		doneBR: make(chan struct{}),
		red:    red,
		blue:   blue,
		rbt:    rbt,
		brt:    brt,
	}
}

func RunSimpleSwitch(red, blue net.Conn, rb, br Translator) {
	NewSimpleSwitch(red, blue, rb, br).Run()
}

func (this *SimpleSwitch) pipe(r, w net.Conn, t Translator, done, stop chan struct{}) {
	defer close(done)
	doneTranslate := make(chan struct{})
	go func() {
		defer close(doneTranslate)
		t.Translate(r, w)
	}()
	select {
	case <-stop:
		return
	case <-doneTranslate:
		return
	}
}

func (this *SimpleSwitch) pipeRB() {
	this.pipe(this.red, this.blue, this.rbt, this.doneRB, this.doneBR)
}

func (this *SimpleSwitch) pipeBR() {
	this.pipe(this.blue, this.red, this.brt, this.doneBR, this.doneRB)
}

func (this *SimpleSwitch) Run() {
	go this.pipeRB()
	go this.pipeBR()
	<-this.doneRB
	<-this.doneBR
}
