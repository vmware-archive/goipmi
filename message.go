// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"bytes"
	"encoding/binary"
	"errors"

	"io"
)

type ipmiSession struct {
	AuthType  uint8
	Sequence  uint32
	SessionID uint32
}

type ipmiHeader struct {
	MsgLen     uint8
	RsAddr     uint8
	NetFnRsLUN uint8
	Checksum   uint8
	RqAddr     uint8
	RqSeq      uint8
	Command
}

// NetworkFunction identifies the functional class of an IPMI message
type NetworkFunction uint8

// Network Function Codes per section 5.1
var (
	NetworkFunctionChassis = NetworkFunction(0x00)
	NetworkFunctionApp     = NetworkFunction(0x06)
)

// General errors
var (
	ErrShortPacket   = errors.New("ipmi: short packet")
	ErrInvalidPacket = errors.New("ipmi: invalid packet")
)

var (
	ipmiHeaderSize  = binary.Size(ipmiHeader{})
	ipmiSessionSize = binary.Size(ipmiSession{})
	ipmiBufSize     = 1024
)

// Message encapsulates an IPMI message
type Message struct {
	*rmcpHeader
	*ipmiSession
	AuthCode [16]byte
	*ipmiHeader
	Data      []byte
	RequestID string
}

// NetFn returns the NetworkFunction portion of the NetFn/RsLUN field
func (m *Message) NetFn() NetworkFunction {
	return NetworkFunction(m.NetFnRsLUN >> 2)
}

// CompletionCode of an IPMI command response
func (m *Message) CompletionCode() CompletionCode {
	return CompletionCode(m.Data[0])
}

// Response specific to the request IPMI command
func (m *Message) Response(data Response) error {
	if m.CompletionCode() != CommandCompleted {
		return m.CompletionCode()
	}
	return binary.Read(bytes.NewBuffer(m.Data), binary.BigEndian, data)
}

func messageFromBytes(buf []byte) (*Message, error) {
	if len(buf) < rmcpHeaderSize+ipmiSessionSize+ipmiHeaderSize {
		return nil, ErrShortPacket
	}

	m := &Message{
		rmcpHeader:  &rmcpHeader{},
		ipmiSession: &ipmiSession{},
		ipmiHeader:  &ipmiHeader{},
	}
	reader := bytes.NewReader(buf)

	if err := binary.Read(reader, binary.BigEndian, m.rmcpHeader); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, m.ipmiSession); err != nil {
		return nil, err
	}
	if m.AuthType != 0 {
		if err := binary.Read(reader, binary.BigEndian, &m.AuthCode); err != nil {
			return nil, err
		}
	}
	if err := binary.Read(reader, binary.BigEndian, m.ipmiHeader); err != nil {
		return nil, err
	}
	if m.headerChecksum() != m.Checksum {
		return nil, ErrInvalidPacket
	}

	if m.MsgLen <= 0 {
		return nil, ErrInvalidPacket
	}
	dataLen := int(m.MsgLen) - ipmiHeaderSize
	data := make([]byte, dataLen+1)
	_, err := reader.Read(data)
	if err != nil {
		return nil, err
	}
	m.Data = data[:dataLen]
	if m.payloadChecksum(m.Data) != data[dataLen] {
		return nil, ErrInvalidPacket
	}

	return m, nil
}

func (m *Message) toBytes(data interface{}) []byte {
	buf := new(bytes.Buffer)

	binaryWrite(buf, m.rmcpHeader)
	binaryWrite(buf, m.ipmiSession)
	if m.AuthType != 0 {
		binaryWrite(buf, m.AuthCode)
	}

	m.MsgLen = uint8(ipmiHeaderSize + binary.Size(data))
	m.Checksum = m.headerChecksum()
	binaryWrite(buf, m.ipmiHeader)

	dlen := buf.Len()
	binaryWrite(buf, data)
	binaryWrite(buf, m.payloadChecksum(buf.Bytes()[dlen:]))

	return buf.Bytes()
}

func (m *Message) headerChecksum() uint8 {
	return checksum(m.RsAddr, m.NetFnRsLUN)
}

func (m *Message) payloadChecksum(data []byte) uint8 {
	return checksum(m.RqAddr, m.RqSeq, uint8(m.Command)) + checksum(data...)
}

func checksum(b ...uint8) uint8 {
	var c uint8
	for _, x := range b {
		c += x
	}
	return -c
}

func binaryWrite(writer io.Writer, data interface{}) {
	err := binary.Write(writer, binary.BigEndian, data)
	if err != nil {
		// shouldn't happen to a bytes.Buffer
		panic(err)
	}
}
