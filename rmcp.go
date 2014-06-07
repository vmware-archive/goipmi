// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

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
