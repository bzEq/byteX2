// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"sync"
)

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

const TLS_APP_PROTO = "com.wandu.spider"
const DEFAULT_HOST = "www.wandu.com"

func CreateBarebonesTLSConfig() (c *tls.Config, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return
	}
	c = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{TLS_APP_PROTO},
	}
	return
}
