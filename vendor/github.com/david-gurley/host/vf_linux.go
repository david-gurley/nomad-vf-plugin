package host

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vishvananda/netlink"
)

func UnbindDriver(vfAddress string) error {
	file := fmt.Sprintf(pciDeviceUnbindDriver, vfAddress)
	return WriteFile(file, []byte(vfAddress))
}

func (vf *Vf) UnbindDriver() error {
	return UnbindDriver(vf.Address)

}

// make sure IOUMMU is enabled in kernel via grub:
// GRUB_CMDLINE_LINUX_DEFAULT="intel_iommu=on iommu=pt vfio_iommu_type1.allow_unsafe_interrupts=1"
func (vf *Vf) BindVfio() error {
	err := vf.UnbindDriver()
	if err != nil {
		return err
	}
	newIDString := fmt.Sprintf("%x %x", vf.VendorIDInt(), vf.DeviceIDInt())
	return WriteFile(vfioNewIdPath, []byte(newIDString))
}

func (vf *Vf) BindHostDriver() error {
	err := vf.UnbindDriver()
	if err != nil {
		return err
	}
	bindPath := fmt.Sprintf(pciDriverBindPath, "iavf")
	// bind to host driver
	err = WriteFile(bindPath, []byte(vf.Address))
	if err != nil {
		return err
	}
	return nil
}

func (vfs *Vfs) GetAllocations() error {
	vfioAllocations, err := VfioAllocations()
	if err != nil {
		return err
	}
	for _, vf := range *vfs {
		vf.Allocated = IsAllocated(vfioAllocations, vf.IommuGroup)
	}
	return nil
}

func GetVfs() (Vfs, error) {
	vfs := Vfs{}
	files, err := os.ReadDir(pciDevicesPath)
	if err != nil {
		return vfs, err
	}
	hostname, err := GetHostname()
	if err != nil {
		return vfs, err
	}
	hostId, err := ReadFileString(hostIdPath)
	if err != nil {
		return vfs, err
	}
	for _, file := range files {
		vf, err := GetVf(filepath.Base(file.Name()))
		if err != nil {
			continue
		}
		vf.Hostname = hostname
		vf.HostID = hostId
		vfs = append(vfs, &vf)
	}
	err = vfs.GetAllocations()
	if err != nil {
		return vfs, err
	}
	return vfs, nil
}

func GetVf(address string) (Vf, error) {
	var vf Vf
	if IsEthernet(address) && IsVf(address) {
		vf.Address = address
		vf.IommuGroup = IommuGroup(vf.Address)
		vf.Driver = Driver(vf.Address)
		vf.PfAddress = VfPf(vf.Address).Address

		// interface name
		d := fmt.Sprintf(pciNetDevicePath, vf.Address)
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			files, err := os.ReadDir(d)
			if err != nil {
				return vf, err
			}
			if len(files) > 1 {
				return vf, errors.New("more than one network device for vf")
			}
			vf.InterfaceName = filepath.Base(files[0].Name())
			if vf.InterfaceName != "" {
				err := vf.GetMacAddress()
				if err != nil {
					return vf, err
				}
				err = vf.GetIPAddresses()
				if err != nil {
					return vf, err
				}
			}
		}

		// PCI Vendor
		vendorId, err := GetPciDeviceVendor(vf.Address)
		if err != nil {
			vf.Vendor = "unknown"
		}
		if pciVendorMap[vendorId] != "" {
			vf.Vendor = pciVendorMap[vendorId]
		} else {
			vf.Vendor = vendorId
		}
		vf.VendorID = vendorId
		// PCI Vendor Device
		deviceId, err := GetPciDeviceDevice(vf.Address)
		if err != nil {
			vf.Device = "unknown"
		}
		if pciDeviceMap[vendorId][deviceId] != "" {
			vf.Device = pciDeviceMap[vendorId][deviceId]
		} else {
			vf.Device = deviceId
		}
		vf.DeviceID = deviceId
	} else {
		return vf, errors.New(fmt.Sprintf("pci device is not an ethernet vf: %s\n", address))
	}
	return vf, nil
}

// get the PF for a given VF address
func VfPf(address string) Pf {
	pfAddr, err := filepath.EvalSymlinks(fmt.Sprintf(pfPath, address))
	if err != nil {
		return Pf{}
	}
	return PfFromVfAddress(filepath.Base(pfAddr))
}

func (vf *Vf) GetMacAddress() error {
	link, err := netlink.LinkByName(vf.InterfaceName)
	if err != nil {
		return err
	}
	vf.MacAddress = link.Attrs().HardwareAddr.String()
	return nil
}

func (vf *Vf) GetIPAddresses() error {
	link, err := netlink.LinkByName(vf.InterfaceName)
	if err != nil {
		return err
	}
	// 0 is for all addresses families
	addresses, err := netlink.AddrList(link, 0)
	if err != nil {
		return err
	}
	for _, addr := range addresses {
		vf.IPAddresses = append(vf.IPAddresses, addr.IPNet.String())

	}
	return nil
}

func (vf *Vf) IPAddressesJoin() string {
	return strings.Join(vf.IPAddresses, ",")
}
