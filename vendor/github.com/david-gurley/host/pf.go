package host

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	pfKindList []string = []string{
		"pci",
	}
)

type PfConfig struct {
	Address string `json:"address"`
	NumVfs  int    `json:"num_vfs"`
	Vfio    bool   `json:"vfio"`
}

type Pf struct {
	Hostname         string   `json:"hostname"`
	HostID           string   `json:"host_id"`
	Address          string   `json:"address"`
	Vendor           string   `json:"vendor"`
	VendorID         string   `json:"vendor_id"`
	Device           string   `json:"device"`
	DeviceID         string   `json:"device_id"`
	HasIPAddress     bool     `json:"has_ip_address"`
	IPAddresses      []string `json:"ip_addresses"` // ip_addr/netmask
	MacAddress       string   `json:"mac_address"`
	Driver           string   `json:"driver"`
	DriverVersion    string   `json:"driver_version"`
	FwVersion        string   `json:"fw_version"`
	InterfaceName    string   `json:"interface_name"`
	File             string   `json:"file"`
	BondMember       uint16   `json:"bond_member"`
	HasSubinterfaces bool     `json:"has_subinterfaces"` // could be dot1q, macvlan, etc.
	TotalVfs         int      `json:"total_vfs"`
	NumVfs           int      `json:"num_vfs"`
	Vfs              Vfs      `json:"vfs"`
}

func GetPfsMap() (map[string]*Pf, error) {
	pfsMap := make(map[string]*Pf)
	pfs, err := GetPfs()
	if err != nil {
		return pfsMap, err
	}
	for _, pf := range pfs {
		pfsMap[pf.Address] = pf
	}
	return pfsMap, nil
}

func (pf *Pf) ID() string {
	return fmt.Sprintf("%s:%s", pf.HostID, pf.Address)
}

func InPfKindList(file string) bool {
	for _, pfKind := range pfKindList {
		match, err := regexp.Match(strings.ToLower(pfKind), []byte(strings.ToLower(file)))
		if err != nil {
			continue
		}
		if match {
			return true
		}

	}
	return false
}

type Pfs []*Pf
