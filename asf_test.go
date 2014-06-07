// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestASFMessageFromBytes(t *testing.T) {
	buf := []byte{}
	_, err := asfMessageFromBytes(buf)
	assert.Error(t, err)
	assert.Equal(t, ErrShortPacket, err)

	buf = make([]byte, rmcpHeaderSize+asfHeaderSize)
	_, err = asfMessageFromBytes(buf)
	assert.NoError(t, err)
}
