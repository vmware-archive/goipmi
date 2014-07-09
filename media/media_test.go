// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMedia(t *testing.T) {
	tests := []struct {
		id ipmi.OemID
		ok bool
	}{
		{ipmi.OemHP, true},
		{ipmi.OemDell, true},
		{ipmi.OemBroadcom, false},
	}

	for _, test := range tests {
		c := &ipmi.Connection{}
		id := &ipmi.DeviceIDResponse{ManufacturerID: test.id}
		_, err := New(c, id)
		if test.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
