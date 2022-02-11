package host

import (
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

type Vf struct {
	PfAddress     string   `json:"pf_address"`
	Hostname      string   `json:"hostname"`
	HostID        string   `json:"host_id"`
	Address       string   `json:"address"`
	Vendor        string   `json:"vendor"`
	VendorID      string   `json:"vendor_id"`
	Device        string   `json:"device"`
	DeviceID      string   `json:"device_id"`
	MacAddress    string   `json:"mac_address"`
	IPAddresses   []string `json:"ip_addresses"`
	Driver        string   `json:"driver"`
	HostDriver    string   `json:"host_driver"`
	InterfaceName string   `json:"interface_name"`
	IommuGroup    string   `json:"iommu_group"`
	Allocated     bool     `json:"allocated"`
}

func (vf *Vf) ID() string {
	return fmt.Sprintf("%s:%s", vf.HostID, vf.Address)
}

func (vf *Vf) Bus() string {
	addr := FromString(vf.Address)
	return addr.Bus
}
func (vf *Vf) VendorIDInt() *big.Int {
	vendorId := new(big.Int)
	fmt.Sscan(vf.VendorID, vendorId)
	return vendorId
}

func (vf *Vf) DeviceIDInt() *big.Int {
	deviceId := new(big.Int)
	fmt.Sscan(vf.DeviceID, deviceId)
	return deviceId
}

type Vfs []*Vf

func (vfs *Vfs) ByPfAddress(pfAddress string) Vfs {
	var pfVfs Vfs
	for _, vf := range *vfs {
		if vf.PfAddress == pfAddress {
			pfVfs = append(pfVfs, vf)
		}
	}
	return pfVfs
}

// Get the next available vf by vendor. will allocate evenly across pfs
func (vfs *Vfs) NextVf(vendorName string) *Vf {
	if len(*vfs) == 0 {
		return &Vf{}
	}
	allocatedBusMap := make(map[string]int)
	allocatedBusList := make(BusCountList, 2)
	freeBusMap := make(map[string][]*Vf)
	for _, vf := range *vfs {
		match, err := regexp.Match(strings.ToLower(vendorName), []byte(strings.ToLower(vf.Vendor)))
		if err != nil {
			continue
		}
		if match {
			if vf.Allocated {
				allocatedBusMap[vf.Bus()]++

			} else {
				allocatedBusMap[vf.Bus()] = allocatedBusMap[vf.Bus()]
				freeBusMap[vf.Bus()] = append(freeBusMap[vf.Bus()], vf)
			}
		}
	}
	i := 0
	for bus, count := range allocatedBusMap {
		allocatedBusList[i] = BusCount{Bus: bus, Count: count}
		i++
	}
	sort.Sort(allocatedBusList)
	return freeBusMap[allocatedBusList[0].Bus][0]
}

func (vfs *Vfs) NumPfs() int {
	pfsMap := make(map[string]int)
	for _, vf := range *vfs {
		pfsMap[vf.PfAddress]++
	}
	count := 0
	for _, _ = range pfsMap {
		count++
	}
	return count
}

// Get the next available vf by pf address
func (vfs *Vfs) NextVfByPf(pfAddress string) *Vf {
	if len(*vfs) == 0 {
		return &Vf{}
	}
	for _, vf := range *vfs {
		pfMatch, err := regexp.Match(strings.ToLower(pfAddress), []byte(strings.ToLower(vf.PfAddress)))
		if err != nil {
			continue
		}
		if vf.Allocated {
			continue

		} else if pfMatch {
			return vf
		}
	}
	return &Vf{}
}

func (vfs *Vfs) ByVendors(vendorNames []string) Vfs {
	var response Vfs
	for _, vf := range *vfs {
		for _, vendorName := range vendorNames {
			match, err := regexp.Match(strings.ToLower(vendorName), []byte(strings.ToLower(vf.Vendor)))
			if err != nil {
				continue
			}
			if match {
				response = append(response, vf)
			}
		}
	}
	return response

}

func GetVfsVendor(vendorName string) (Vfs, error) {
	allVfs, err := GetVfs()
	var vfs Vfs
	if err != nil {
		return vfs, err
	}
	for _, vf := range allVfs {
		match, err := regexp.Match(strings.ToLower(vendorName), []byte(strings.ToLower(vf.Vendor)))
		if err != nil {
			continue
		}
		if match {
			vfs = append(vfs, vf)
		}
	}
	return vfs, nil
}

func IsVf(address string) bool {
	return !IsPf(address)
}
