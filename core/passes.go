// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bytes"
	"compress/gzip"
	"crypto/rc4"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
)

type DummyPass struct{}

func (this *DummyPass) RunOnBytes(p []byte) ([]byte, error) {
	return p, nil
}

type CopyPass struct{}

func (this *CopyPass) RunOnBytes(p []byte) ([]byte, error) {
	c := make([]byte, len(p))
	copy(c, p)
	return c, nil
}

type RC4Pass struct {
	C *rc4.Cipher
}

func (this *RC4Pass) RunOnBytes(p []byte) ([]byte, error) {
	result := make([]byte, len(p))
	this.C.XORKeyStream(result, p)
	return result, nil
}

const EMOJI_BEGIN = "😀😁😂😃😄😅😆"

const EMOJI_END = "😐😑😶😏😔"

type Padder struct{}

func (this *Padder) RunOnBytes(p []byte) ([]byte, error) {
	r := bytes.NewBuffer([]byte(EMOJI_BEGIN))
	r.Write(p)
	r.Write([]byte(EMOJI_END))
	return r.Bytes(), nil
}

type Unpadder struct{}

func (this *Unpadder) RunOnBytes(p []byte) ([]byte, error) {
	b := []byte(EMOJI_BEGIN)
	e := []byte(EMOJI_END)
	if len(p) < len(b)+len(e) {
		return p, errors.New("Unavailable buffer length")
	}
	return p[len(b) : len(p)-len(e)], nil
}

type Base64Enc struct{}

func (this *Base64Enc) RunOnBytes(p []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(p)), nil
}

type Base64Dec struct{}

func (this *Base64Dec) RunOnBytes(p []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(p))
}

type Compressor struct{}

func (this *Compressor) RunOnBytes(p []byte) (result []byte, err error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	defer zw.Close()
	if _, err = zw.Write(p); err == nil {
		err = zw.Flush()
	}
	result = buf.Bytes()
	return
}

type Decompressor struct{}

func (this *Decompressor) RunOnBytes(p []byte) (result []byte, err error) {
	buf := bytes.NewBuffer(p)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		return buf.Bytes(), err
	}
	defer zr.Close()
	result, err = ioutil.ReadAll(zr)
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		return
	}
	return result, nil
}
