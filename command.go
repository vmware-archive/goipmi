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

// Command fields on an IPMI message
type Command uint8

// Command Number Assignments (table G-1)
const (
	CommandGetDeviceID              = Command(0x01)
	CommandGetAuthCapabilities      = Command(0x38)
	CommandGetSessionChallenge      = Command(0x39)
	CommandActivateSession          = Command(0x3a)
	CommandSetSessionPrivilegeLevel = Command(0x3b)
	CommandCloseSession             = Command(0x3c)
	CommandChassisControl           = Command(0x02)
	CommandChassisStatus            = Command(0x01)
	CommandSetSystemBootOptions     = Command(0x08)
	CommandGetSystemBootOptions     = Command(0x09)
	CommandGetMcIdString            = Command(0x09)
	CommandSetMcIdString            = Command(0x0a)
)

// Request structure
type Request struct {
	NetworkFunction
	Command
	Data interface{}
}

// Response to an IPMI request must include at least a CompletionCode
type Response interface {
	Code() uint8
}

// DeviceIDRequest per section 20.1
type DeviceIDRequest struct{}

// DeviceIDResponse per section 20.1
type DeviceIDResponse struct {
	CompletionCode
	DeviceID                uint8
	DeviceRevision          uint8
	FirmwareRevision1       uint8
	FirmwareRevision2       uint8
	IPMIVersion             uint8
	AdditionalDeviceSupport uint8
	ManufacturerID          OemID
	ProductID               uint16
}

// AuthCapabilitiesRequest per section 22.13
type AuthCapabilitiesRequest struct {
	ChannelNumber uint8
	PrivLevel     uint8
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

// AuthType
const (
	AuthTypeNone = iota
	AuthTypeMD2
	AuthTypeMD5
	authTypeReserved
	AuthTypePassword
	AuthTypeOEM
)

// PrivLevel
const (
	PrivLevelNone = iota
	PrivLevelCallback
	PrivLevelUser
	PrivLevelOperator
	PrivLevelAdmin
	PrivLevelOEM
)

// SessionChallengeRequest per section 22.16
type SessionChallengeRequest struct {
	AuthType uint8
	Username [16]uint8
}

// SessionChallengeResponse per section 22.16
type SessionChallengeResponse struct {
	CompletionCode
	TemporarySessionID uint32
	Challenge          [16]byte
}

// ActivateSessionRequest per section 22.17
type ActivateSessionRequest struct {
	AuthType  uint8
	PrivLevel uint8
	AuthCode  [16]uint8
	InSeq     [4]uint8
}

// ActivateSessionResponse per section 22.17
type ActivateSessionResponse struct {
	CompletionCode
	AuthType   uint8
	SessionID  uint32
	InboundSeq uint32
	MaxPriv    uint8
}

// SessionPrivilegeLevelRequest per section 22.18
type SessionPrivilegeLevelRequest struct {
	PrivLevel uint8
}

// SessionPrivilegeLevelResponse per section 22.18
type SessionPrivilegeLevelResponse struct {
	CompletionCode
	NewPrivilegeLevel uint8
}

// CloseSessionRequest per section 22.19
type CloseSessionRequest struct {
	SessionID uint32
}

// CloseSessionResponse per section 22.19
type CloseSessionResponse struct {
	CompletionCode
}
