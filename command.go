// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import "fmt"

// Command fields on an IPMI message
type Command uint8

// Command Number Assignments (table G-1)
var (
	CommandGetDeviceID              = Command(0x01)
	CommandGetAuthCapabilities      = Command(0x38)
	CommandGetSessionChallenge      = Command(0x39)
	CommandActivateSession          = Command(0x3a)
	CommandSetSessionPrivilegeLevel = Command(0x3b)
	CommandCloseSession             = Command(0x3c)
	CommandChassisControl           = Command(0x02)
	CommandSetSystemBootOptions     = Command(0x08)
)

// CompletionCode is the first byte in the data field of all IPMI responses
type CompletionCode uint8

// Code returns the CompletionCode as uint8
func (c CompletionCode) Code() uint8 {
	return uint8(c)
}

// Error for CompletionCode
func (c CompletionCode) Error() string {
	return fmt.Sprintf("Completion Code: %d", c)
}

// Completion Codes per section 5.2
var (
	CommandCompleted       = CompletionCode(0x00)
	InvalidCommand         = CompletionCode(0xc1)
	DestinationUnavailable = CompletionCode(0xd3)
	UnspecifiedError       = CompletionCode(0xff)
)

// Response to an IPMI request must include at least a CompletionCode
type Response interface {
	Code() uint8
}

// DeviceIDResponse per section 20.1
type DeviceIDResponse struct {
	CompletionCode
	DeviceID                uint8
	DeviceRevision          uint8
	FirmwareRevision1       uint8
	FirmwareRevision2       uint8
	IPMIVersion             uint8
	AdditionalDeviceSupport uint8
	ManufacturerID          uint16
	ProductID               uint16
}

// AuthCapabilitiesResponse per section 22.13
type AuthCapabilitiesResponse struct {
	CompletionCode
	ChannelNumber   uint8
	AuthTypeSupport uint8
	Status          uint8
	Reserved        uint8
	OEMID           uint16
	OEMAux          uint8
}

// AuthTypeSupport
const (
	AuthTypeNone = (1 << iota)
	AuthTypeMD2
	AuthTypeMD5
	AuthTypeReserved
	AuthTypePassword
	AuthTypeOEM
)

// SessionChallengeResponse per section 22.16
type SessionChallengeResponse struct {
	CompletionCode
	TemporarySessionID uint32
	Challenge          [15]byte
}

// ActivateSessionResponse per section 22.17
type ActivateSessionResponse struct {
	CompletionCode
	AuthType   uint8
	SessionID  uint32
	InboundSeq uint32
	MaxPriv    uint8
}

// SessionPrivilegeLevelResponse per section 22.18
type SessionPrivilegeLevelResponse struct {
	CompletionCode
	NewPrivilegeLevel uint8
}
