// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRMCPHeaderFromBytes(t *testing.T) {
	buf := []byte{}
	_, err := rmcpHeaderFromBytes(buf)
	assert.Error(t, err)
	assert.Equal(t, ErrShortPacket, err)

	buf = make([]byte, rmcpHeaderSize)
	_, err = rmcpHeaderFromBytes(buf)
	assert.NoError(t, err)
}
