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
	"encoding/binary"
	"fmt"
)

const (
	rmcpClassASF  = 0x06
	rmcpClassIPMI = 0x07
	rmcpVersion1  = 0x06
)

var (
	rmcpHeaderSize = binary.Size(rmcpHeader{})
)

type rmcpHeader struct {
	Version            uint8
	Reserved           uint8
	RMCPSequenceNumber uint8
	Class              uint8
}

func rmcpHeaderFromBytes(buf []byte) (*rmcpHeader, error) {
	if len(buf) < rmcpHeaderSize {
		return nil, ErrShortPacket
	}
	return &rmcpHeader{
		buf[0],
		buf[1],
		buf[2],
		buf[3],
	}, nil
}

func (h *rmcpHeader) unsupportedClass() error {
	return fmt.Errorf("unsupported RMCP class: %d", h.Class)
}
