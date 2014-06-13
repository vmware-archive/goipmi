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
			Param: 0x05,
		},
	}
	raw := requestToStrings(req)
	assert.Equal(t, []string{"0x00", "0x09", "0x05", "0x00", "0x00"}, raw)
}

func TestBootFlagsParse(t *testing.T) {
	res := &SystemBootFlagsResponse{}
	err := responseFromString("01 05 80 3c 00 00 00", res)
	assert.NoError(t, err)
	assert.Equal(t, BootDeviceFloppy, res.BootDeviceSelector())
}
