// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

// A byte transfer based on socks5 protocol.

package main

import (
	"flag"
	"log"
	"net"

	"github.com/bzEq/byteX2/core"
	socks5 "github.com/bzEq/byteX2/socks"
)

var options struct {
	Key    string
	Local  string
	Remote string
	Server string
	Dummy  bool
}

func startServer(addr string, handler func(net.Conn)) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go handler(c)
	}
}

const RC4_KEY = "0xcafec0de"

func createPackPass() core.Pass {
	pm := core.NewPassManager()
	// Compress
	pm.AddPass(&core.Compressor{})
	return pm
}

func createUnpackPass() core.Pass {
	pm := core.NewPassManager()
	// Decompress
	pm.AddPass(&core.Decompressor{})
	return pm
}

func handleLocal(red net.Conn) {
	defer red.Close()
	blue, err := net.Dial("tcp", options.Remote)
	if err != nil {
		log.Println(err)
		return
	}
	defer blue.Close()
	var rb, br core.Translator
	if options.Dummy {
		rb = &core.Repeater{}
		br = &core.Repeater{}
	} else {
		rb = &core.HTTPPacker{
			P: createPackPass(),
		}
		br = &core.HTTPUnpacker{
			P: createUnpackPass(),
		}
	}
	core.RunSimpleSwitch(red, blue, rb, br)
}

func handle(red net.Conn) {
	defer red.Close()
	blue := core.MakePipe()
	go func() {
		defer blue[0].Close()
		rb := &core.HTTPUnpacker{
			P: createUnpackPass(),
		}
		br := &core.HTTPPacker{
			P: createPackPass(),
		}
		core.RunSimpleSwitch(red, blue[0], rb, br)
	}()
	server := &socks5.Server{}
	server.Serve(blue[1])
}

func main() {
	flag.StringVar(&options.Key, "key", "0xc0de", "Key for the cipher")
	flag.StringVar(&options.Local, "c", ":1080", "Client side server")
	flag.StringVar(&options.Remote, "r", "", "Remote server address")
	flag.StringVar(&options.Server, "s", ":8010", "Server")
	flag.BoolVar(&options.Dummy, "d", false, "Just a dummy repeater")
	flag.Parse()
	if options.Remote != "" {
		startServer(options.Local, handleLocal)
	} else {
		startServer(options.Server, handle)
	}
}
