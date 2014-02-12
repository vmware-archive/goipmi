// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

type ChassisControl uint8
type BootDevice uint8

const (
	netfnChassis = 0x0

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

type ChassisStatus struct {
	PowerState        uint8
	LastPowerEvent    uint8
	State             uint8
	FrontControlPanel uint8
}

type BootFlags struct {
	BootDeviceSelector BootDevice
}

func (s *ChassisStatus) IsSystemPowerOn() bool {
	return (s.PowerState & SystemPower) == SystemPower
}

func (s *ChassisStatus) String() string {
	if s.IsSystemPowerOn() {
		return "on"
	} else {
		return "off"
	}
}

func (s *ChassisStatus) PowerRestorePolicy() uint8 {
	return (s.PowerState & 0x60) >> 5
}

func (s *ChassisStatus) request() []byte {
	return []byte{
		netfnChassis,
		0x1,
	}
}

func (s *ChassisStatus) parse(data []byte) {
	s.PowerState = data[0]
	s.LastPowerEvent = data[1]
	s.State = data[2]
	// optional per ipmi spec
	if len(data) > 3 {
		s.FrontControlPanel = data[3]
	}
}

func getBootParamRequest(id uint8) []byte {
	return []byte{
		netfnChassis,
		0x9,
		id & 0x7f,
		0x0,
		0x0,
	}
}

func (b *BootFlags) request() []byte {
	return getBootParamRequest(0x5)
}

func (b *BootFlags) parse(data []byte) {
	b.BootDeviceSelector = BootDevice((data[3] >> 2) & 0x0f)
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
