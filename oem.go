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

// OemID aka IANA assigned Enterprise Number per:
// http://www.iana.org/assignments/enterprise-numbers/enterprise-numbers
// Note that constants defined here are the same subset that ipmitool recognizes.
type OemID uint16

// IANA assigned manufacturer IDs
const (
	OemUnknown              = OemID(0)
	OemHP                   = OemID(11)
	OemSun                  = OemID(42)
	OemNokia                = OemID(94)
	OemBull                 = OemID(107)
	OemHitachi116           = OemID(116)
	OemNEC                  = OemID(119)
	OemToshiba              = OemID(186)
	OemIntel                = OemID(343)
	OemTatung               = OemID(373)
	OemHitachi399           = OemID(399)
	OemDell                 = OemID(674)
	OemLMC                  = OemID(2168)
	OemRadiSys              = OemID(4337)
	OemBroadcom             = OemID(4413)
	OemMagnum               = OemID(5593)
	OemTyan                 = OemID(6653)
	OemNewisys              = OemID(9237)
	OemFujitsuSiemens       = OemID(10368)
	OemAvocent              = OemID(10418)
	OemPeppercon            = OemID(10437)
	OemSupermicro           = OemID(10876)
	OemOSA                  = OemID(11102)
	OemGoogle               = OemID(11129)
	OemPICMG                = OemID(12634)
	OemRaritan              = OemID(13742)
	OemKontron              = OemID(15000)
	OemPPS                  = OemID(16394)
	OemAMI                  = OemID(20974)
	OemNokiaSiemensNetworks = OemID(28458)
	OemSupermicro47488      = OemID(47488)
)

var oemStrings = map[OemID]string{
	OemUnknown:              "Unknown",
	OemHP:                   "Hewlett-Packard",
	OemSun:                  "Sun Microsystems",
	OemNokia:                "Nokia",
	OemBull:                 "Bull Company",
	OemHitachi116:           "Hitachi",
	OemNEC:                  "NEEC",
	OemToshiba:              "Toshiba",
	OemIntel:                "Intel Corporation",
	OemTatung:               "Tatung",
	OemHitachi399:           "Hitachi",
	OemDell:                 "Dell Inc",
	OemLMC:                  "LMC",
	OemRadiSys:              "RadiSys Corporation",
	OemBroadcom:             "Broadcom Corporation",
	OemMagnum:               "Magnum Technologies",
	OemTyan:                 "Tyan Computer Corporation",
	OemNewisys:              "Newisys",
	OemFujitsuSiemens:       "Fujitsu Siemens",
	OemAvocent:              "Avocent",
	OemPeppercon:            "Peppercon AG",
	OemSupermicro:           "Supermicro",
	OemOSA:                  "OSA",
	OemGoogle:               "Google",
	OemPICMG:                "PICMG",
	OemRaritan:              "Raritan",
	OemKontron:              "Kontron",
	OemPPS:                  "Pigeon Point Systems",
	OemAMI:                  "AMI",
	OemNokiaSiemensNetworks: "Nokia Siemens Networks",
	OemSupermicro47488:      "Supermicro",
}

func (id OemID) String() string {
	if s, ok := oemStrings[id]; ok {
		return s
	}
	return fmt.Sprintf("Unknown (%d)", id)
}
