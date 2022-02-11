package host

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/safchain/ethtool"
)

// Prefer PCI addres versus interface name as interface could be
// managed by userspace.
const (
	pciDevicesPath        = "/sys/bus/pci/devices"
	pciDeviceClassPath    = "/sys/bus/pci/devices/%s/class"
	pciDeviceVendorPath   = "/sys/bus/pci/devices/%s/vendor"
	pciDeviceDevicePath   = "/sys/bus/pci/devices/%s/device"
	pciNetDevicePath      = "/sys/bus/pci/devices/%s/net"
	pciDeviceUnbindDriver = "/sys/bus/pci/devices/%s/driver/unbind"
	pciDriverPath         = "/sys/bus/pci/devices/%s/driver"
	iommuGroupPath        = "/sys/bus/pci/devices/%s/iommu_group"
	driverPath            = "/sys/bus/pci/devices/%s/driver"
	vfioDevPath           = "/dev/vfio/%v"
	pciDriverBindPath     = "/sys/bus/pci/drivers/%s/bind"
	vfioNewIdPath         = "/sys/bus/pci/drivers/vfio-pci/new_id"
)

func GetPciDeviceClass(address string) (string, error) {
	file := fmt.Sprintf(pciDeviceClassPath, address)
	return ReadFileString(file)
}
func GetPciDeviceVendor(address string) (string, error) {
	file := fmt.Sprintf(pciDeviceVendorPath, address)
	return ReadFileString(file)
}
func GetPciDeviceDevice(address string) (string, error) {
	file := fmt.Sprintf(pciDeviceDevicePath, address)
	return ReadFileString(file)
}

func InterfaceNameFromAddress(address string) string {
	interfaceName := ""
	files, err := os.ReadDir(pfNetDevPath)
	if err != nil {
		return ""
	}
	// Ignoring errors after here because process could have ended
	for _, file := range files {
		dev := fmt.Sprintf("%v/%v", pfNetDevPath, file.Name())
		s, err := filepath.EvalSymlinks(dev)
		if err != nil {
			continue
		}
		match, err := regexp.Match(strings.ToLower(address), []byte(strings.ToLower(s)))
		if err != nil {
			continue
		}
		if match {
			interfaceName = file.Name()
			break
		}
	}
	return interfaceName
}

func KernelDeviceDriver(interfaceName string) (DriverInfo, error) {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		return DriverInfo{}, err
	}
	defer ethHandle.Close()
	driverInfo, err := ethHandle.DriverInfo(interfaceName)
	if err != nil {
		return DriverInfo{}, err
	}
	return DriverInfo{
		driverInfo.Driver,
		driverInfo.Version,
		driverInfo.FwVersion,
	}, nil
}

// is the pci device controlled by vfio?
func IsVfio(address string) bool {
	f, err := filepath.EvalSymlinks(fmt.Sprintf(pciDriverPath, address))
	if err != nil {
		return false
	}

	if f == "/sys/bus/pci/drivers/vfio-pci" {
		return true
	}
	return false
}

// get the iommu group for a given pci bus address
func IommuGroup(address string) string {
	s, err := filepath.EvalSymlinks(fmt.Sprintf(iommuGroupPath, address))
	if err != nil {
		return ""
	}
	return filepath.Base(s)
}

// get the driver for a give pci bus address
func Driver(address string) string {
	d, err := filepath.EvalSymlinks(fmt.Sprintf(driverPath, address))
	if err != nil {
		return ""
	}
	return filepath.Base(d)
}

// is the given iommu group allocated to a process?
func IsAllocated(allocations []string, iommuGroup string) bool {
	for _, allocation := range allocations {
		match := fmt.Sprintf(vfioDevPath, iommuGroup)
		if allocation == match {
			return true
		}
	}
	return false
}

// Find all current vfio allocations
// part of this from mitchellh/go-ps/process_unix.go
func VfioAllocations() ([]string, error) {
	allocations := make([]string, 0)
	d, err := os.Open("/proc")
	if err != nil {
		return allocations, err
	}
	defer d.Close()

	for {
		names, err := d.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return allocations, err
		}

		for _, name := range names {
			// We only care if the name starts with a numeric
			if name[0] < '0' || name[0] > '9' {
				continue
			}
			files, err := os.ReadDir(fmt.Sprintf("/proc/%v/fd", name))
			if err != nil {
				continue
			}
			// Ignoring errors after here because process could have ended
			for _, file := range files {
				fd := fmt.Sprintf("/proc/%v/fd/%v", name, file.Name())
				s, err := filepath.EvalSymlinks(fd)
				if err != nil {
					continue
				}
				r := regexp.MustCompile("/dev/vfio/[0-9]+")
				match := r.Match([]byte(strings.ToLower(s)))
				if match {
					allocations = append(allocations, s)
				}

			}

		}
	}
	return allocations, nil
}
