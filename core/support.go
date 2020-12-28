// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"io"
	"net"
	"sync"
)

const TLS_APP_PROTO = "com.wandu.spider"
const DEFAULT_HOST = "www.wandu.com"

type OnceCloser struct {
	once sync.Once
	c    io.Closer
}

func NewOnceCloser(c io.Closer) io.Closer {
	return &OnceCloser{c: c}
}

func (this *OnceCloser) Close() error {
	this.once.Do(func() { this.c.Close() })
	return nil
}

func MakePipe() (pipe [2]net.Conn) {
	pipe[0], pipe[1] = net.Pipe()
	return
}
