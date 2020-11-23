// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package socks

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/bzEq/byteX2/core"
)

const VER = 5

const (
	CMD_CONNECT = iota + 1
	CMD_BIND
	CMD_UDP_ASSOCIATE
)

const (
	ATYP_IPV4 = iota + 1
	_
	ATYP_DOMAINNAME
	ATYP_IPV6
)

const (
	REP_SUCC = iota
	REP_GENERAL_SERVER_FAILURE
	REP_CONNECTION_NOT_ALLOWED
	REP_NETWORK_UNREACHABLE
	REP_HOST_UNREACHABLE
	REP_CONNECTION_REFUSED
	REP_TTL_EXPIRED
	REP_COMMAND_NOT_SUPPORTED
	REP_ADDRESS_TYPE_NOT_SUPPORTED
	REP_UNASSIGNED_START
)

type Server struct{}

type Request struct {
	VER, CMD, ATYP byte
	DST_ADDR       []byte
	DST_PORT       [2]byte
}

type Reply struct {
	VER, REP, ATYP byte
	BND_ADDR       []byte
	BND_PORT       [2]byte
}

func (this *Server) exchangeMetadata(rw io.ReadWriter) (err error) {
	buf := make([]byte, 255)
	// VER, NMETHODS.
	if _, err = io.ReadFull(rw, buf[:2]); err != nil {
		return
	}
	// METHODS.
	methods := buf[1]
	if _, err = io.ReadFull(rw, buf[:methods]); err != nil {
		return
	}
	// No auth for now.
	if _, err = rw.Write([]byte{VER, 0}); err != nil {
		return
	}
	return
}

func (this *Server) receiveRequest(r io.Reader) (req Request, err error) {
	buf := make([]byte, net.IPv6len)
	// VER, CMD, RSV, ATYP
	if _, err = io.ReadFull(r, buf[:4]); err != nil {
		return req, err
	}
	req.VER = buf[0]
	req.CMD = buf[1]
	req.ATYP = buf[3]
	switch req.ATYP {
	case ATYP_IPV6:
		if _, err = io.ReadFull(r, buf[:net.IPv6len]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, net.IPv6len)
		copy(req.DST_ADDR, buf[:net.IPv6len])
	case ATYP_IPV4:
		if _, err = io.ReadFull(r, buf[:net.IPv4len]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, net.IPv4len)
		copy(req.DST_ADDR, buf[:net.IPv4len])
	case ATYP_DOMAINNAME:
		if _, err = io.ReadFull(r, buf[:1]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, buf[0])
		if _, err = io.ReadFull(r, req.DST_ADDR); err != nil {
			return
		}
	default:
		return req, fmt.Errorf("Unsupported ATYP: %d", req.ATYP)
	}
	_, err = io.ReadFull(r, buf[:2])
	if err != nil {
		return req, err
	}
	copy(req.DST_PORT[:2], buf[:2])
	return req, nil
}

func (this *Server) getDialAddress(req Request) string {
	port := binary.BigEndian.Uint16(req.DST_PORT[:2])
	switch req.ATYP {
	case ATYP_IPV6:
		return fmt.Sprintf("%s:%d", "["+net.IP(req.DST_ADDR).String()+"]", port)
	case ATYP_IPV4:
		return fmt.Sprintf("%s:%d", net.IP(req.DST_ADDR).String(), port)
	case ATYP_DOMAINNAME:
		return fmt.Sprintf("%s:%d", string(req.DST_ADDR), port)
	default:
		return ""
	}
}

func (this *Server) sendReply(r Reply, w io.Writer) (err error) {
	// FIXME: Respect Reply.
	if _, err = w.Write([]byte{r.VER, r.REP, 0, r.ATYP}); err != nil {
		return
	}
	if _, err = w.Write(r.BND_ADDR); err != nil {
		return
	}
	if _, err = w.Write(r.BND_PORT[:2]); err != nil {
		return
	}
	return
}

func (this *Server) Serve(c net.Conn) (err error) {
	defer c.Close()
	if err = this.exchangeMetadata(c); err != nil {
		return
	}
	req, err := this.receiveRequest(c)
	if err != nil {
		return
	}
	// FIXME: Strictly follow RFC.
	reply := Reply{
		VER:  VER,
		REP:  REP_SUCC,
		ATYP: req.ATYP,
	}
	if req.VER != VER {
		err = fmt.Errorf("Unsupported SOCKS version: %v", req.VER)
		return
	}
	if req.CMD != CMD_CONNECT {
		err = fmt.Errorf("Unsupported CMD: %d", req.CMD)
		reply.REP = REP_COMMAND_NOT_SUPPORTED
		this.sendReply(reply, c)
		return
	}
	if err = this.sendReply(reply, c); err != nil {
		return
	}
	addr := this.getDialAddress(req)
	remoteConn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	defer remoteConn.Close()
	core.RunSimpleSwitch(c, remoteConn, &core.Repeater{}, &core.Repeater{})
	return nil
}
