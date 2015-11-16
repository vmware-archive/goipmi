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

type ChassisControl uint8
type BootDevice uint8

const (
	ControlPowerDown      = ChassisControl(0x0)
	ControlPowerUp        = ChassisControl(0x1)
	ControlPowerCycle     = ChassisControl(0x2)
	ControlPowerHardReset = ChassisControl(0x3)
	ControlPowerPulseDiag = ChassisControl(0x4)
	ControlPowerAcpiSoft  = ChassisControl(0x5)

	BootDeviceNone          = BootDevice(0x00)
	BootDevicePxe           = BootDevice(0x04)
	BootDeviceDisk          = BootDevice(0x08)
	BootDeviceSafe          = BootDevice(0x0c)
	BootDeviceDiag          = BootDevice(0x10)
	BootDeviceCdrom         = BootDevice(0x14)
	BootDeviceBios          = BootDevice(0x18)
	BootDeviceRemoteFloppy  = BootDevice(0x1c)
	BootDeviceRemotePrimary = BootDevice(0x24)
	BootDeviceRemoteCdrom   = BootDevice(0x20)
	BootDeviceRemoteDisk    = BootDevice(0x2c)
	BootDeviceFloppy        = BootDevice(0x3c)

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

	BootParamSetInProgress = 0x0
	BootParamSvcPartSelect = 0x1
	BootParamSvcPartScan   = 0x2
	BootParamFlagValid     = 0x3
	BootParamInfoAck       = 0x4
	BootParamBootFlags     = 0x5
	BootParamInitInfo      = 0x6
	BootParamInitMbox      = 0x7
)

// ChassisStatusRequest per section 28.2
type ChassisStatusRequest struct{}

// ChassisStatusResponse per section 28.2
type ChassisStatusResponse struct {
	CompletionCode
	PowerState        uint8
	LastPowerEvent    uint8
	State             uint8
	FrontControlPanel uint8
}

// ChassisControlRequest per section 28.3
type ChassisControlRequest struct {
	ChassisControl
}

// ChassisControlResponse per section 28.3
type ChassisControlResponse struct {
	CompletionCode
}

// SetSystemBootOptionsRequest per section 28.12
type SetSystemBootOptionsRequest struct {
	Param uint8
	Data  []uint8
}

// SetSystemBootOptionsResponse per section 28.12
type SetSystemBootOptionsResponse struct {
	CompletionCode
}

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

// MarshalBinary implementation to handle variable length Data
func (r *SetSystemBootOptionsRequest) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 1+len(r.Data))
	buf[0] = r.Param
	copy(buf[1:], r.Data)
	return buf, nil
}

var validSetBootOptionsDataLength = map[uint8]int{
	BootParamInfoAck:   2,
	BootParamBootFlags: 5,
}

// UnmarshalBinary implementation to handle variable length Data
func (r *SetSystemBootOptionsRequest) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return ErrShortPacket
	}
	r.Param = buf[0]
	r.Data = buf[1:]
	if l, ok := validSetBootOptionsDataLength[r.Param]; ok {
		if len(r.Data) < l {
			return ErrShortPacket
		}
	}
	return nil
}

// MarshalBinary implementation to handle variable length Data
func (r *SystemBootOptionsResponse) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3+len(r.Data))
	buf[0] = byte(r.CompletionCode)
	buf[1] = r.Version
	buf[2] = r.Param
	copy(buf[3:], r.Data)
	return buf, nil
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

// UnmarshalBinary implementation to handle variable length Data
func (r *ChassisStatusResponse) UnmarshalBinary(buf []byte) error {
	if len(buf) < 4 {
		return ErrShortPacket
	}
	r.CompletionCode = CompletionCode(buf[0])
	r.PowerState = buf[1]
	r.LastPowerEvent =  buf[2]
	r.State = buf[3]
	if len(buf) > 4 {
		r.FrontControlPanel = buf[4]
	} else {
		r.FrontControlPanel = 0
	}
	return nil
}

func (r *SystemBootOptionsResponse) BootDeviceSelector() BootDevice {
	return BootDevice(((r.Data[1] >> 2) & 0x0f) << 2)
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

var bootDeviceStrings = map[BootDevice]string{
	BootDeviceNone:          "none",
	BootDevicePxe:           "pxe",
	BootDeviceDisk:          "disk",
	BootDeviceSafe:          "safe",
	BootDeviceDiag:          "diag",
	BootDeviceCdrom:         "cdrom",
	BootDeviceBios:          "bios",
	BootDeviceRemoteFloppy:  "rfloppy",
	BootDeviceRemotePrimary: "rprimary",
	BootDeviceRemoteCdrom:   "rcdrom",
	BootDeviceRemoteDisk:    "rdisk",
	BootDeviceFloppy:        "floppy",
}

func (d BootDevice) String() string {
	if s, ok := bootDeviceStrings[d]; ok {
		return s
	}
	return "unknown"
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
