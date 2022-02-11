package host

import (
	"regexp"
	"strings"
)

var (
	pciEthernetDeviceClasses []string = []string{
		"0x020000",
	}
	pciVendorMap map[string]string = map[string]string{
		"0x1dd8": "pensando",
		"0x8086": "intel",
		"0x15b3": "mellanox",
		"0x14e4": "broadcom",
	}
	pciDeviceMap map[string]map[string]string = map[string]map[string]string{
		"0x1dd8": map[string]string{
			"0x1000": "DSC Capri Upstream Port",
			"0x1001": "DSC Virtual Downstream Port",
			"0x1002": "DSC Ethernet Controller",
			"0x1003": "DSC Ethernet Controller VF",
			"0x1004": "DSC Management Controller",
			"0x1007": "DSC Storage Accelerator",
		},
		"0x8086": map[string]string{
			"0x10CA": "82576 Virtual Function",
			"0x1520": "I350 Virtual Function",
			"0x1521": "I350 Gigabit Network Connection",
			"0x37cd": "x722 Virtual Function",
			"0x37d2": "Ethernet Connection x722 for 10GBase-T",
			"0x37d0": "Ethernet Connection x722 for SFP",
			"0x10fb": "82599ES 10-Gigabit SFI/SFP+ Network Connection",
			"0x1572": "Ethernet Controller X710 for 10GbE SFP+",
		},
		"0x15b3": map[string]string{
			"0x1017": "MT27640 Family [ConnectX-5]",
		},
		"0x14e4": map[string]string{
			"0x1682": "NetXtreme BCM57762 Gigabit Ethernet PCIe",
		},
	}
)

func IsEthernet(address string) bool {
	c, err := GetPciDeviceClass(address)
	if err != nil {
		return false
	}
	for _, cl := range pciEthernetDeviceClasses {
		if c == cl {
			return true
		}
	}
	return false

}

type DriverInfo struct {
	Name            string `json:"name"`
	DriverVersion   string `json:"driver_version"`
	FirmwareVersion string `json:"firmware_version"`
}

func BusTrimmed(address string) string {
	a := FromString(address)
	leadingZero, err := regexp.Compile("^0[0-9]+")
	if err != nil {
		return ""
	}
	if leadingZero.MatchString(a.Bus) {
		return a.Bus[1:]
	}
	return a.Bus
}

type BusCount struct {
	Bus   string
	Count int
}
type BusCountList []BusCount

func (b BusCountList) Len() int           { return len(b) }
func (b BusCountList) Less(i, j int) bool { return b[i].Count < b[j].Count }
func (b BusCountList) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type Address struct {
	Domain   string
	Bus      string
	Slot     string
	Function string
}

// String() returns the canonical [D]BSF representation of this Address
func (addr *Address) String() string {
	if addr.Domain != "" && addr.Bus != "" && addr.Slot != "" && addr.Function != "" {
		return addr.Domain + ":" + addr.Bus + ":" + addr.Slot + "." + addr.Function
	}
	return ""
}

// Given a string address, returns a complete Address struct, filled in with
// domain, bus, slot and function components. The address string may either
// be in $BUS:$SLOT.$FUNCTION (BSF) format or it can be a full PCI address
// that includes the 4-digit $DOMAIN information as well:
// $DOMAIN:$BUS:$SLOT.$FUNCTION.
//
func FromString(address string) *Address {
	addrLowered := strings.ToLower(address)
	matches := regexAddress.FindStringSubmatch(addrLowered)
	if len(matches) == 6 {
		dom := "0000"
		if matches[1] != "" {
			dom = matches[2]
		}
		return &Address{
			Domain:   dom,
			Bus:      matches[3],
			Slot:     matches[4],
			Function: matches[5],
		}
	}
	return &Address{}
}
func FromNetString(address string) *Address {
	addrLowered := strings.ToLower(address)
	matches := regexNetAddress.FindStringSubmatch(addrLowered)
	if len(matches) == 6 {
		dom := "0000"
		if matches[1] != "" {
			dom = matches[2]
		}
		return &Address{
			Domain:   dom,
			Bus:      matches[3],
			Slot:     matches[4],
			Function: matches[5],
		}
	}
	return &Address{}

}
