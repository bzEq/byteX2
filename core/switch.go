// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"net"
)

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

func (this *SimpleSwitch) pipe(r, w net.Conn, t Translator, done, stop chan struct{}) error {
	defer close(done)
	return t.Translate(r, w, stop)
}

func (this *SimpleSwitch) pipeRB() error {
	return this.pipe(this.red, this.blue, this.rbt, this.doneRB, this.doneBR)
}

func (this *SimpleSwitch) pipeBR() error {
	return this.pipe(this.blue, this.red, this.brt, this.doneBR, this.doneRB)
}

func (this *SimpleSwitch) Run() {
	go this.pipeRB()
	go this.pipeBR()
	<-this.doneRB
	<-this.doneBR
}
