// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

type ChassisControl uint8
type BootDevice uint8

const (
	ControlPowerDown      = ChassisControl(0x0)
	ControlPowerUp        = ChassisControl(0x1)
	ControlPowerCycle     = ChassisControl(0x2)
	ControlPowerHardReset = ChassisControl(0x3)
	ControlPowerPulseDiag = ChassisControl(0x4)
	ControlPowerAcpiSoft  = ChassisControl(0x5)

	BootDeviceNone   = BootDevice(0)
	BootDevicePxe    = BootDevice(1)
	BootDeviceDisk   = BootDevice(2)
	BootDeviceSafe   = BootDevice(3)
	BootDeviceDiag   = BootDevice(4)
	BootDeviceCdrom  = BootDevice(5)
	BootDeviceBios   = BootDevice(6)
	BootDeviceFloppy = BootDevice(15)

	SystemPower       = 0x1
	PowerOverload     = 0x2
	PowerInterlock    = 0x4
	MainPowerFault    = 0x8
	PowerControlFault = 0x10

	PowerRestorePolicyAlwaysOff = 0x0
	PowerRestorePolicyPrevious  = 0x1
	PowerRestorePolicyAlwaysOn  = 0x2
	PowerRestorePolicyUnknown   = 0x3

	PowerEventUnknown   = 0x0
	PowerEventAcFailed  = 0x1
	PowerEventOverload  = 0x2
	PowerEventInterlock = 0x4
	PowerEventFault     = 0x8
	PowerEventCommand   = 0x10

	ChassisIntrusion  = 0x1
	FrontPanelLockout = 0x2
	DriveFault        = 0x4
	CoolingFanFault   = 0x8

	SleepButtonDisable  = 0x80
	DiagButtonDisable   = 0x40
	ResetButtonDisable  = 0x20
	PowerButtonDisable  = 0x10
	SleepButtonDisabled = 0x08
	DiagButtonDisabled  = 0x04
	ResetButtonDisabled = 0x02
	PowerButtonDisabled = 0x01
)

// ChassisStatusRequest per section 28.3
type ChassisStatusRequest struct{}

// ChassisStatusResponse per section 28.3
type ChassisStatusResponse struct {
	CompletionCode
	PowerState        uint8
	LastPowerEvent    uint8
	State             uint8
	FrontControlPanel uint8
}

// SetSystemBootOptionsRequest per section 28.12
type SetSystemBootOptionsRequest struct {
	Param uint8
	Data  []uint8
}

// SetSystemBootOptionsResponse per section 28.12
type SetSystemBootOptionsResponse struct{}

// SystemBootOptionsRequest per section 28.13
type SystemBootOptionsRequest struct {
	Param uint8
	Set   uint8
	Block uint8
}

// SystemBootOptionsResponse per section 28.13
type SystemBootOptionsResponse struct {
	CompletionCode
	Version uint8
	Param   uint8
	Data    []uint8
}

// UnmarshalBinary implementation to handle variable length Data
func (r *SystemBootOptionsResponse) UnmarshalBinary(buf []byte) error {
	if len(buf) < 3 {
		return ErrShortPacket
	}
	r.CompletionCode = CompletionCode(buf[0])
	r.Version = buf[1]
	r.Param = buf[2]
	r.Data = buf[3:]
	return nil
}

func (r *SystemBootOptionsResponse) BootDeviceSelector() BootDevice {
	return BootDevice((r.Data[1] >> 2) & 0x0f)
}

func (s *ChassisStatusResponse) IsSystemPowerOn() bool {
	return (s.PowerState & SystemPower) == SystemPower
}

func (s *ChassisStatusResponse) String() string {
	if s.IsSystemPowerOn() {
		return "on"
	}
	return "off"
}

func (s *ChassisStatusResponse) PowerRestorePolicy() uint8 {
	return (s.PowerState & 0x60) >> 5
}

func (d BootDevice) String() string {
	switch d {
	case BootDeviceNone:
		return "none"
	case BootDevicePxe:
		return "pxe"
	case BootDeviceDisk:
		return "disk"
	case BootDeviceSafe:
		return "safe"
	case BootDeviceDiag:
		return "diag"
	case BootDeviceCdrom:
		return "cdrom"
	case BootDeviceBios:
		return "bios"
	case BootDeviceFloppy:
		return "floppy"
	}
	panic("unknown device")
}

func (c ChassisControl) String() string {
	switch c {
	case ControlPowerDown:
		return "down"
	case ControlPowerUp:
		return "up"
	case ControlPowerCycle:
		return "cycle"
	case ControlPowerHardReset:
		return "reset"
	case ControlPowerPulseDiag:
		return "diag"
	case ControlPowerAcpiSoft:
		return "acpi"
	}
	panic("unknown control")
}
