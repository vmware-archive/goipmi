// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMedia(t *testing.T) {
	tests := []struct {
		id  ipmi.OemID
		ok  bool
		dev ipmi.BootDevice
	}{
		{ipmi.OemHP, true, ipmi.BootDeviceNone},
		{ipmi.OemDell, true, ipmi.BootDeviceRemoteFloppy},
		{ipmi.OemBroadcom, false, 0},
	}

	for _, test := range tests {
		c := &ipmi.Connection{}
		id := &ipmi.DeviceIDResponse{ManufacturerID: test.id}
		m, err := New(c, id)
		if test.ok {
			assert.NoError(t, err)
			assert.Equal(t, test.dev, m.BootDevice("foo.img"))
		} else {
			assert.Error(t, err)
		}
	}
}
