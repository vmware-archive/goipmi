// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
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
