// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package floppy

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloppy(t *testing.T) {
	f, err := Create(&File{
		Name: "foo",
		Data: strings.NewReader("yo."),
	})
	assert.NoError(t, err)
	defer os.Remove(f)
	assert.NoError(t, err)
}
