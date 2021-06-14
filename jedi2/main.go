// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

// A byte transfer based on socks5 protocol.

package main

import (
	"flag"
	"log"
	"net"

	"github.com/bzEq/byteX2/core"
	socks5 "github.com/bzEq/byteX2/socks5"
)

var options struct {
	Local       string
	Next        string
	Transparent bool
}

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

func startRelayer() {
	l, err := net.Listen("tcp", options.Local)
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
		if options.Next != "" {
			go serveAsIntermediateRelayer(c)
		} else {
			go serveAsEndRelayer(c)
		}
	}
}

func serveAsEndRelayer(red net.Conn) {
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
	err := server.Serve(blue[1])
	if err != nil {
		log.Println(err)
	}
}

func serveAsIntermediateRelayer(red net.Conn) {
	defer red.Close()
	blue, err := net.Dial("tcp", options.Next)
	if err != nil {
		log.Println(err)
		return
	}
	defer blue.Close()
	var rb, br core.Translator
	if options.Transparent {
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

func main() {
	flag.StringVar(&options.Local, "c", ":1080", "Address of local relayer")
	flag.StringVar(&options.Next, "r", "", "Address of next-hop relayer")
	flag.BoolVar(&options.Transparent, "t", false, "Indicate transparent relayer")
	flag.Parse()
	startRelayer()
}
