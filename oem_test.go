// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOEM(t *testing.T) {
	assert.Equal(t, "Dell Inc", OemDell.String())
	assert.Equal(t, "Hewlett-Packard", OemHP.String())
}
