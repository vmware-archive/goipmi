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
	"encoding/binary"
	"fmt"
	"io"
)

const (
	asfMessageTypePing = 0x80
	asfMessageTypePong = 0x40
	asfIANA            = 0x000011be
)

var (
	asfHeaderSize = binary.Size(asfHeader{})
)

type asfHeader struct {
	IANAEnterpriseNumber uint32
	MessageType          uint8
	MessageTag           uint8
	Reserved             uint8
	DataLength           uint8
}

type asfMessage struct {
	*rmcpHeader
	*asfHeader
	Data []byte
}

type asfPong struct {
	IANAEnterpriseNumber  uint32
	OEM                   uint32
	SupportedEntities     uint8
	SupportedInteractions uint8
	Reserved              [6]uint8
}

func (h *asfHeader) unsupportedMessageType() error {
	return fmt.Errorf("unsupported ASF message type: %d", h.MessageType)
}

func asfMessageFromBytes(buf []byte) (*asfMessage, error) {
	hlen := rmcpHeaderSize + asfHeaderSize
	if len(buf) < hlen {
		return nil, ErrShortPacket
	}

	rh, err := rmcpHeaderFromBytes(buf)
	if err != nil {
		return nil, err
	}

	buf = buf[rmcpHeaderSize:]
	ah := &asfHeader{
		binary.BigEndian.Uint32(buf),
		buf[4],
		buf[5],
		buf[6],
		buf[7],
	}

	return &asfMessage{
		rmcpHeader: rh,
		asfHeader:  ah,
		Data:       buf[asfHeaderSize:],
	}, nil
}

func (m *asfMessage) toBytes(data interface{}) []byte {
	buf := new(bytes.Buffer)

	m.binaryWrite(buf, m.rmcpHeader)
	m.binaryWrite(buf, m.asfHeader)
	if data != nil {
		m.binaryWrite(buf, data)
	}

	return buf.Bytes()
}

// Response specific to the request ASF command
func (m *asfMessage) response(data interface{}) error {
	return binary.Read(bytes.NewBuffer(m.Data), binary.BigEndian, data)
}

func (m *asfMessage) binaryWrite(writer io.Writer, data interface{}) {
	err := binary.Write(writer, binary.BigEndian, data)
	if err != nil {
		// shouldn't happen to a bytes.Buffer
		panic(err)
	}
}

func (m *asfPong) valid() bool {
	return m.SupportedEntities&0x80 != 0
}
