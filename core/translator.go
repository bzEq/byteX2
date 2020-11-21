// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Translator interface {
	Translate(in, out net.Conn, stop chan struct{}) error
}

type Repeater struct{}

const DEFAULT_COPY_SIZE = 4096

func (this *Repeater) Translate(in, out net.Conn, stop chan struct{}) error {
	for {
		select {
		case <-stop:
			break
		default:
			in.SetReadDeadline(time.Now().Add(DEFAULT_TIMEOUT * time.Second))
			out.SetReadDeadline(time.Now().Add(2 * DEFAULT_TIMEOUT * time.Second))
			_, err := io.CopyN(out, in, DEFAULT_COPY_SIZE)
			if err != nil {
				return err
			}
		}
	}
}

type HTTPPacker struct {
	P Pass
}

const HTTP_BUFFER_SIZE = 1 << 16
const DEFAULT_TIMEOUT = 600

func (this *HTTPPacker) Translate(in, out net.Conn, stop chan struct{}) error {
	buf := make([]byte, HTTP_BUFFER_SIZE)
	for {
		select {
		case <-stop:
			break
		default:
			in.SetReadDeadline(time.Now().Add(DEFAULT_TIMEOUT * time.Second))
			n, err := in.Read(buf)
			if err != nil {
				return err
			}
			body, err := this.P.RunOnBytes(buf[:n])
			if err != nil {
				return err
			}
			req, err := http.NewRequest("POST", "/", bytes.NewReader(body))
			if err != nil {
				return err
			}
			req.Host = DEFAULT_HOST
			req.Header.Add("User-Agent", TLS_APP_PROTO)
			out.SetWriteDeadline(time.Now().Add(DEFAULT_TIMEOUT * time.Second))
			err = req.Write(out)
			if err != nil {
				return err
			}
		}
	}
}

type HTTPUnpacker struct {
	P Pass
}

func (this *HTTPUnpacker) Translate(in, out net.Conn, stop chan struct{}) error {
	b := bufio.NewReader(in)
	for {
		select {
		case <-stop:
			break
		default:
			in.SetReadDeadline(time.Now().Add(DEFAULT_TIMEOUT * time.Second))
			req, err := http.ReadRequest(b)
			if err != nil {
				return err
			}
			req.Host = DEFAULT_HOST
			req.Header.Add("User-Agent", TLS_APP_PROTO)
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}
			if body, err = this.P.RunOnBytes(body); err != nil {
				return err
			}
			out.SetWriteDeadline(time.Now().Add(DEFAULT_TIMEOUT * time.Second))
			_, err = out.Write(body)
			if err != nil {
				return err
			}
		}
	}
}
