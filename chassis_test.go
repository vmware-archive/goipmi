/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChassisStatusRequest(t *testing.T) {
	req := &Request{
		NetworkFunctionChassis,
		CommandChassisStatus,
		&ChassisStatusRequest{},
	}
	raw := requestToStrings(req)
	assert.Equal(t, []string{"0x00", "0x01"}, raw)
}

func TestChassisStatusParse(t *testing.T) {
	status := &ChassisStatusResponse{}
	err := responseFromString("21 10 40 54", status)
	assert.NoError(t, err)

	assert.Equal(t, true, status.IsSystemPowerOn())
	assert.Equal(t, uint8(SystemPower), status.PowerState&SystemPower)
	assert.Equal(t, uint8(0x0), status.PowerState&PowerOverload)

	assert.Equal(t, uint8(PowerRestorePolicyPrevious), status.PowerRestorePolicy())

	assert.Equal(t, uint8(PowerEventCommand), status.LastPowerEvent&PowerEventCommand)
	assert.Equal(t, uint8(0x0), status.LastPowerEvent&PowerEventAcFailed)

	assert.Equal(t, uint8(0x0), status.FrontControlPanel&SleepButtonDisable)
	assert.Equal(t, uint8(DiagButtonDisabled), status.FrontControlPanel&DiagButtonDisabled)

	assert.Equal(t, uint8(0x0), status.State&CoolingFanFault)
}

func TestBootFlagsRequest(t *testing.T) {
	req := &Request{
		NetworkFunctionChassis,
		CommandGetSystemBootOptions,
		&SystemBootOptionsRequest{
			Param: BootParamBootFlags,
		},
	}
	raw := requestToStrings(req)
	assert.Equal(t, []string{"0x00", "0x09", "0x05", "0x00", "0x00"}, raw)
}

func TestBootFlagsParse(t *testing.T) {
	res := &SystemBootOptionsResponse{}
	err := responseFromString("01 05 80 3c 00 00 00", res)
	assert.NoError(t, err)
	assert.Equal(t, BootDeviceFloppy, res.BootDeviceSelector())
}
