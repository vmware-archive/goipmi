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

import "fmt"

// CompletionCode is the first byte in the data field of all IPMI responses
type CompletionCode uint8

// Completion Codes per section 5.2
const (
	CommandCompleted     = CompletionCode(0x00)
	ErrNodeBusy          = CompletionCode(0xc0)
	ErrInvalidCommand    = CompletionCode(0xc1)
	ErrInvalidLunCommand = CompletionCode(0xc2)
	ErrCommandTimeout    = CompletionCode(0xc3)
	ErrOutOfSpace        = CompletionCode(0xc4)
	ErrInvalidResv       = CompletionCode(0xc5)
	ErrDataTruncated     = CompletionCode(0xc6)
	ErrShortPacket       = CompletionCode(0xc7)
	ErrLongPacket        = CompletionCode(0xc8)
	ErrParamRange        = CompletionCode(0xc9)
	ErrRequestData       = CompletionCode(0xca)
	ErrNoObj             = CompletionCode(0xcb)
	ErrInvalidPacket     = CompletionCode(0xcc)
	ErrInvalidObjCommand = CompletionCode(0xcd)
	ErrNoResponse        = CompletionCode(0xce)
	ErrDuplicateRequest  = CompletionCode(0xcf)
	ErrRepoUpMode        = CompletionCode(0xd0)
	ErrFirmwareUpMode    = CompletionCode(0xd1)
	ErrInitMode          = CompletionCode(0xd2)
	ErrDestUnavail       = CompletionCode(0xd3)
	ErrPrivLevel         = CompletionCode(0xd4)
	ErrInvalidState      = CompletionCode(0xd5)
	ErrUnspecified       = CompletionCode(0xff)
)

var completionCodes = map[CompletionCode]string{
	CommandCompleted:     "Command completed normally",
	ErrNodeBusy:          "Node busy",
	ErrInvalidCommand:    "Unrecognized or unsupported command",
	ErrInvalidLunCommand: "Command invalid for given LUN",
	ErrCommandTimeout:    "Timeout while processing command",
	ErrOutOfSpace:        "Out of space",
	ErrInvalidResv:       "Reservation canceled or invalid reservation ID",
	ErrDataTruncated:     "Request data truncated",
	ErrShortPacket:       "Request data length invalid",
	ErrLongPacket:        "Request data field length limit exceeded",
	ErrParamRange:        "Parameter out of range",
	ErrRequestData:       "Cannot return number of requested data bytes",
	ErrNoObj:             "Requested sensor, data, or record not present",
	ErrInvalidPacket:     "Invalid data field in request",
	ErrInvalidObjCommand: "Command illegal for specified sensor or record type",
	ErrNoResponse:        "Command response could not be provided",
	ErrDuplicateRequest:  "Cannot execute duplicated request",
	ErrRepoUpMode:        "SDR repository in update mode",
	ErrFirmwareUpMode:    "Device in firmware update mode",
	ErrInitMode:          "BMC initialization or initialization agent running",
	ErrDestUnavail:       "Destination unavailable",
	ErrPrivLevel:         "Insufficient privilege level",
	ErrInvalidState:      "Command or param not supported in present state",
	ErrUnspecified:       "Unspecified error",
}

// Code returns the CompletionCode as uint8
func (c CompletionCode) Code() uint8 {
	return uint8(c)
}

// Error for CompletionCode
func (c CompletionCode) Error() string {
	if s, ok := completionCodes[c]; ok {
		return s
	}
	return fmt.Sprintf("Completion Code: %X", uint8(c))
}
