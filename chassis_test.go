package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChassisStatusRequest(t *testing.T) {
	status := &ChassisStatus{}
	raw := rawEncode(status.request())
	assert.Equal(t, []string{"0x00", "0x01"}, raw)
}

func TestChassisStatusParse(t *testing.T) {
	status := &ChassisStatus{}
	data := rawDecode("21 10 40 54")
	status.parse(data)

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
	flags := &BootFlags{}
	raw := rawEncode(flags.request())
	assert.Equal(t, []string{"0x00", "0x09", "0x05", "0x00", "0x00"}, raw)
}

func TestBootFlagsParse(t *testing.T) {
	data := rawDecode("01 05 80 3c 00 00 00")
	flags := &BootFlags{}
	flags.parse(data)
	assert.Equal(t, BootDeviceFloppy, flags.BootDeviceSelector)
}
