// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandError(t *testing.T) {
	assert.Error(t, InvalidCommand)
}
