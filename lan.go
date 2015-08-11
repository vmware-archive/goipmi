/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type lan struct {
	*Connection
	ipmiSession
	rqSeq    uint8
	conn     net.Conn
	active   bool
	authcode [16]uint8
	username [16]uint8
	priv     uint8
	lun      uint8
	timeout  time.Duration
}

func newLanTransport(c *Connection) transport {
	l := &lan{Connection: c}

	copy(l.username[:], c.Username[:])
	copy(l.authcode[:], c.Password[:])

	return l
}

func (l *lan) dial() (net.Conn, error) {
	// TODO: support more than just udp4
	addr := net.JoinHostPort(l.Hostname, strconv.Itoa(l.Port))
	return net.Dial("udp4", addr)
}

func (l *lan) open() error {
	conn, err := l.dial()
	if err != nil {
		return err
	}
	l.conn = conn

	// TODO: options
	l.priv = PrivLevelAdmin
	l.timeout = time.Second * 5
	l.lun = 0

	return l.openSession()
}

func (l *lan) close() error {
	if l.active {
		err := l.closeSession()
		if err != nil {
			log.Printf("error closing session: %s", err)
		}
		l.active = false
	}

	if l.conn != nil {
		_ = l.conn.Close()
		l.conn = nil
	}

	return nil
}

func (l *lan) send(req *Request, res Response) error {
	err := l.sendPacket(l.message(req))
	if err != nil {
		return err
	}

	m, err := l.recvMessage()
	if err != nil {
		return err
	}

	return m.Response(res)
}

func (*lan) Console() error {
	fmt.Println("Console not supported. Press Enter to continue.")
	r := make([]byte, 1)
	_, err := os.Stdin.Read(r)
	return err
}

func (l *lan) sendPacket(buf []byte) error {
	_, err := l.conn.Write(buf)
	return err
}

func (l *lan) recvPacket() ([]byte, error) {
	buf := make([]byte, ipmiBufSize)

	err := l.conn.SetReadDeadline(time.Now().Add(l.timeout))
	if err != nil {
		return nil, err
	}

	n, err := l.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func (l *lan) recvMessage() (*Message, error) {
	buf, err := l.recvPacket()
	if err != nil {
		return nil, err
	}

	header, err := rmcpHeaderFromBytes(buf)
	if err != nil {
		return nil, err
	}

	if header.Class != rmcpClassIPMI {
		return nil, header.unsupportedClass()
	}

	return messageFromBytes(buf)
}

func (l *lan) nextSequence() uint32 {
	if l.Sequence != 0 {
		l.Sequence++
	}
	return l.Sequence
}

func (l *lan) nextRqSeq() uint8 {
	l.rqSeq++
	return l.rqSeq << 2
}

func (l *lan) message(r *Request) []byte {
	m := &Message{
		rmcpHeader: &rmcpHeader{
			Version:            rmcpVersion1,
			Class:              rmcpClassIPMI,
			RMCPSequenceNumber: 0xff,
		},
		ipmiSession: &ipmiSession{
			Sequence:  l.nextSequence(),
			SessionID: l.SessionID,
		},
		ipmiHeader: &ipmiHeader{
			RsAddr:     0x20, // bmcSlaveAddr
			NetFnRsLUN: uint8(r.NetworkFunction)<<2 | l.lun&3,
			Command:    r.Command,
			RqAddr:     0x81, // remoteSWID
			RqSeq:      l.nextRqSeq(),
		},
	}

	if l.active && l.AuthType != 0 {
		copy(m.AuthCode[:], l.authcode[:])
		m.AuthType = l.AuthType
	}

	msg := m.toBytes(r.Data)

	if l.active && l.AuthType == AuthTypeMD5 {
		hlen := rmcpHeaderSize + ipmiSessionSize
		// offset is location of ipmiHeader.RsAddr
		offset := hlen + len(m.AuthCode) + 1
		// rewrite m.AuthCode field
		md5 := l.authMD5(msg[offset:])
		copy(msg[hlen:], md5)
	}

	return msg
}

// per section 22.17.1
func (l *lan) authMD5(data []uint8) []uint8 {
	h := md5.New()

	binaryWrite(h, l.authcode)
	binaryWrite(h, l.SessionID)
	binaryWrite(h, data)
	binaryWrite(h, l.Sequence)
	binaryWrite(h, l.authcode)

	return h.Sum(nil)
}

func (l *lan) openSession() error {
	if err := l.ping(); err != nil {
		return err
	}

	if err := l.getAuthCapabilities(); err != nil {
		return err
	}

	res, err := l.getSessionChallenge()
	if err != nil {
		return err
	}

	if err := l.activateSession(res); err != nil {
		return err
	}

	return l.setSessionPriv()
}

func (l *lan) ping() error {
	msg := &asfMessage{
		rmcpHeader: &rmcpHeader{
			Version:            rmcpVersion1,
			Class:              rmcpClassASF,
			RMCPSequenceNumber: 0xff,
		},
		asfHeader: &asfHeader{
			IANAEnterpriseNumber: asfIANA,
			MessageType:          asfMessageTypePing,
		},
	}

	if err := l.sendPacket(msg.toBytes(nil)); err != nil {
		return err
	}

	buf, err := l.recvPacket()
	if err != nil {
		return err
	}
	m, err := asfMessageFromBytes(buf)
	if err != nil {
		return err
	}
	if m.MessageType != asfMessageTypePong {
		return m.unsupportedMessageType()
	}

	pong := &asfPong{}
	if err := m.response(pong); err != nil {
		return err
	}
	if !pong.valid() {
		return errors.New("IPMI not supported")
	}

	return nil
}

func (l *lan) getAuthCapabilities() error {
	req := &Request{
		NetworkFunctionApp,
		CommandGetAuthCapabilities,
		AuthCapabilitiesRequest{
			ChannelNumber: 0x0e, // lanChannelE
			PrivLevel:     l.priv,
		},
	}
	res := &AuthCapabilitiesResponse{}

	if err := l.send(req, res); err != nil {
		return err
	}

	for _, t := range []uint8{AuthTypeMD5, AuthTypePassword, AuthTypeNone} {
		if (res.AuthTypeSupport & (1 << t)) != 0 {
			l.AuthType = t
			break
		}
		log.Printf("BMC did not offer a supported AuthType")
		return CompletionCode(0xd4)
	}

	return nil
}

func (l *lan) getSessionChallenge() (*SessionChallengeResponse, error) {
	req := &Request{
		NetworkFunctionApp,
		CommandGetSessionChallenge,
		SessionChallengeRequest{
			AuthType: l.AuthType,
			Username: l.username,
		},
	}
	res := &SessionChallengeResponse{}

	if err := l.send(req, res); err != nil {
		return nil, err
	}

	l.SessionID = res.TemporarySessionID
	return res, nil
}

func (l *lan) inSeq() [4]uint8 {
	seq := [4]uint8{}
	if _, err := rand.Read(seq[:]); err != nil {
		panic(err)
	}
	return seq
}

func (l *lan) activateSession(sc *SessionChallengeResponse) error {
	req := &Request{
		NetworkFunctionApp,
		CommandActivateSession,
		ActivateSessionRequest{
			AuthType:  l.AuthType,
			PrivLevel: l.priv,
			AuthCode:  sc.Challenge,
			InSeq:     l.inSeq(),
		},
	}
	res := &ActivateSessionResponse{}

	l.active = true

	if err := l.send(req, res); err != nil {
		l.active = false
		return err
	}

	l.SessionID = res.SessionID
	l.AuthType = res.AuthType
	l.Sequence = res.InboundSeq

	return nil
}

func (l *lan) setSessionPriv() error {
	req := &Request{
		NetworkFunctionApp,
		CommandSetSessionPrivilegeLevel,
		SessionPrivilegeLevelRequest{
			PrivLevel: l.priv,
		},
	}
	res := &SessionPrivilegeLevelResponse{}

	if err := l.send(req, res); err != nil {
		return err
	}

	l.priv = res.NewPrivilegeLevel

	return nil
}

func (l *lan) closeSession() error {
	req := &Request{
		NetworkFunctionApp,
		CommandCloseSession,
		CloseSessionRequest{
			SessionID: l.SessionID,
		},
	}

	return l.send(req, &CloseSessionResponse{})
}
