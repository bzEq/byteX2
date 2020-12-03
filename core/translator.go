// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Translator interface {
	Translate(in net.Conn, out net.Conn) error
}

const DEFAULT_TIMEOUT = 600

type Repeater struct{}

func (this *Repeater) Translate(in net.Conn, out net.Conn) error {
	_, err := io.Copy(out, in)
	return err
}

type HTTPPacker struct {
	P Pass
}

const HTTP_BUFFER_SIZE = 16 << 10

func (this *HTTPPacker) Translate(in net.Conn, out net.Conn) error {
	buf := make([]byte, HTTP_BUFFER_SIZE)
	for {
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

type HTTPUnpacker struct {
	P Pass
}

func (this *HTTPUnpacker) Translate(in net.Conn, out net.Conn) error {
	b := bufio.NewReader(in)
	for {
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

type LVPacker struct{}

type LVUnpacker struct{}

func (this *LVPacker) Translate(in net.Conn, out net.Conn) error {
	l := make([]byte, binary.MaxVarintLen64)
	b := make([]byte, HTTP_BUFFER_SIZE)
	for {
		n, err := in.Read(b)
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return err
			}
			return nil
		}
		binary.PutUvarint(l, uint64(n))
		if _, err = out.Write(l); err != nil {
			return err
		}
		if _, err = out.Write(b[:n]); err != nil {
			return err
		}
	}
}

func (this *LVUnpacker) Translate(in net.Conn, out net.Conn) error {
	l := make([]byte, binary.MaxVarintLen64)
	for {
		if _, err := io.ReadFull(in, l); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return err
			}
			return nil
		}
		x, n := binary.Uvarint(l)
		if n <= 0 {
			return errors.New("Can't read a valid length")
		}
		b := make([]byte, x)
		if _, err := io.ReadFull(in, b); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return err
			}
			return nil
		}
		if _, err := out.Write(b); err != nil {
			return err
		}
	}
}
