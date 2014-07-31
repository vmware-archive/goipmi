// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"testing"

	"github.com/stretchr/testify/assert"
)

type driver struct{}

func (*driver) Insert(DeviceMap) error {
	return nil
}

func (*driver) Eject() error {
	return nil
}

func TestMedia(t *testing.T) {
	Register(ipmi.OemBull, func(*ipmi.Client) (Driver, error) {
		return &driver{}, nil
	})

	tests := []struct {
		id ipmi.OemID
		ok bool
	}{
		{ipmi.OemBull, true},
		{ipmi.OemBroadcom, false},
	}

	for _, test := range tests {
		c, err := ipmi.NewClient(&ipmi.Connection{Interface: "lan"})
		assert.NoError(t, err)
		_, err = New(c, test.id)
		if test.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
