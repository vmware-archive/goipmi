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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageFromBytes(t *testing.T) {
	buf := make([]byte, rmcpHeaderSize+ipmiHeaderSize)
	_, err := messageFromBytes(buf)
	assert.Error(t, err)
	assert.Equal(t, ErrShortPacket, err)

	buf = make([]byte, ipmiBufSize)
	_, err = messageFromBytes(buf)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPacket, err)
}

func TestChecksum(t *testing.T) {
	buf := make([]byte, ipmiBufSize)
	buf[16] = 0x38
	buf[512] = 0x3c
	c := checksum(buf...)
	assert.Equal(t, uint8(0x0), c+uint8(0x38+0x3c))
}

type testFixedSizeData struct {
	One   uint8
	Two   uint32
	Three uint16
}

func (m *testFixedSizeData) Code() uint8 {
	return m.One
}

type testVariableSizeData struct {
	testFixedSizeData
	Four []byte
}

func (m *testVariableSizeData) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	binaryWrite(buf, &m.testFixedSizeData)
	_, err := buf.Write(m.Four)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *testVariableSizeData) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	err := binary.Read(buf, binary.LittleEndian, &m.testFixedSizeData)
	if err != nil {
		return err
	}
	m.Four = buf.Bytes()
	return nil
}

func TestDataMarshalFixed(t *testing.T) {
	msgIn := &testFixedSizeData{1, 2, 3}
	buf := messageDataToBytes(msgIn)
	msgOut := &testFixedSizeData{}
	err := messageDataFromBytes(buf, msgOut)
	assert.NoError(t, err)
	assert.Equal(t, msgIn, msgOut)
}

func TestDataMarshalVariable(t *testing.T) {
	data := []byte{4, 5, 6}
	msgIn := &testVariableSizeData{testFixedSizeData{1, 2, 3}, data}
	buf := messageDataToBytes(msgIn)
	assert.Equal(t, binary.Size(msgIn.testFixedSizeData)+len(data), len(buf))
	msgOut := &testVariableSizeData{}
	err := messageDataFromBytes(buf, msgOut)
	assert.NoError(t, err)
	assert.Equal(t, msgIn, msgOut)
}
