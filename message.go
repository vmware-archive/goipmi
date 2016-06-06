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
	"bytes"
	"encoding"
	"encoding/binary"

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
	NetworkFunctionDcmi    = NetworkFunction(0x2c)
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

// Request specific to the request IPMI command
// Unmarshal errors are returned as a Response such that they can be
// propagated to the client.
func (m *Message) Request(data interface{}) Response {
	err := messageDataFromBytes(m.Data, data)
	if err != nil {
		if e, ok := err.(CompletionCode); ok {
			return e
		}
		return ErrUnspecified
	}
	return nil
}

// Response specific to the request IPMI command
func (m *Message) Response(data Response) error {
	if m.CompletionCode() != CommandCompleted {
		return m.CompletionCode()
	}
	return messageDataFromBytes(m.Data, data)
}

func messageDataFromBytes(buf []byte, data interface{}) error {
	if decoder, ok := data.(encoding.BinaryUnmarshaler); ok {
		return decoder.UnmarshalBinary(buf)
	}
	return binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, data)
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

	if err := binary.Read(reader, binary.LittleEndian, m.rmcpHeader); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, m.ipmiSession); err != nil {
		return nil, err
	}
	if m.AuthType != 0 {
		if err := binary.Read(reader, binary.LittleEndian, &m.AuthCode); err != nil {
			return nil, err
		}
	}
	if err := binary.Read(reader, binary.LittleEndian, m.ipmiHeader); err != nil {
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

func messageDataToBytes(data interface{}) []byte {
	if encoder, ok := data.(encoding.BinaryMarshaler); ok {
		buf, err := encoder.MarshalBinary()
		if err != nil {
			panic(err)
		}
		return buf
	}
	buf := new(bytes.Buffer)
	binaryWrite(buf, data)
	return buf.Bytes()
}

func (m *Message) toBytes(data interface{}) []byte {
	dbuf := messageDataToBytes(data)
	buf := new(bytes.Buffer)

	binaryWrite(buf, m.rmcpHeader)
	binaryWrite(buf, m.ipmiSession)
	if m.AuthType != 0 {
		binaryWrite(buf, m.AuthCode)
	}

	m.MsgLen = uint8(ipmiHeaderSize + len(dbuf))
	m.Checksum = m.headerChecksum()
	binaryWrite(buf, m.ipmiHeader)

	dlen := buf.Len()
	_, _ = buf.Write(dbuf)
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
	err := binary.Write(writer, binary.LittleEndian, data)
	if err != nil {
		// shouldn't happen to a bytes.Buffer
		panic(err)
	}
}
