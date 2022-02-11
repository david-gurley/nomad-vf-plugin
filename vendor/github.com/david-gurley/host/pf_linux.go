package host

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/safchain/ethtool"
	"github.com/vishvananda/netlink"
)

const (
	checkIfPf      = "/sys/bus/pci/devices/%s/physfn"
	pfPath         = "/sys/bus/pci/devices/%s/physfn"
	pfNumVfsPath   = "/sys/bus/pci/devices/%s/sriov_numvfs"
	pfTotalVfsPath = "/sys/bus/pci/devices/%s/sriov_totalvfs"
	pfNetDevPath   = "/sys/class/net/"
)

func GetPfByAddress(address string) (*Pf, error) {
	pfs, err := GetPfs()
	if err != nil {
		return &Pf{}, err
	}
	for _, pf := range pfs {
		if pf.Address == address {
			return pf, nil
		}
	}
	return &Pf{}, nil
}

func GetPfs() ([]*Pf, error) {
	pfs := make([]*Pf, 0)
	vfs, err := GetVfs()
	if err != nil {
		return pfs, err
	}
	files, err := os.ReadDir(pciDevicesPath)
	if err != nil {
		return pfs, err
	}
	hostname, err := GetHostname()
	if err != nil {
		return pfs, err
	}
	hostId, err := ReadFileString(hostIdPath)
	if err != nil {
		return pfs, err
	}
	for _, file := range files {
		pf, err := GetPf(filepath.Base(file.Name()))
		if err != nil {
			continue
		}
		pf.Hostname = hostname
		pf.HostID = hostId
		pf.Vfs = vfs.ByPfAddress(pf.Address)
		pfs = append(pfs, &pf)

	}
	return pfs, nil
}

// should we look at getting information fron netlink vs. traversing
// directory structures?
func GetPf(address string) (Pf, error) {
	var pf Pf
	if IsEthernet(address) && IsPf(address) {
		pf.Address = address
		d := fmt.Sprintf(pciNetDevicePath, pf.Address)
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			files, err := os.ReadDir(d)
			if err != nil {
				return pf, err
			}
			if len(files) > 1 {
				return pf, errors.New("more than one network device for pf")
			}
			pf.InterfaceName = filepath.Base(files[0].Name())
			pf.IPAddresses = make([]string, 0)
			if pf.InterfaceName != "" {
				err = pf.PopulateLinkAttrs()
				if err != nil {
					return pf, err
				}
			}
		}
		// PCI Vendor
		vendorId, err := GetPciDeviceVendor(pf.Address)
		if err != nil {
			pf.Vendor = "unknown:"
		}
		if pciVendorMap[vendorId] != "" {
			pf.Vendor = pciVendorMap[vendorId]
		} else {
			pf.Vendor = vendorId
		}
		pf.VendorID = vendorId
		// PCI Vendor Device
		deviceId, err := GetPciDeviceDevice(pf.Address)
		if err != nil {
			pf.Device = "unknown"
		}
		if pciDeviceMap[vendorId][deviceId] != "" {
			pf.Device = pciDeviceMap[vendorId][deviceId]
		} else {
			pf.Device = deviceId
		}
		pf.DeviceID = deviceId
		driverInfo, err := KernelDeviceDriver(pf.InterfaceName)
		if err != nil {
			return pf, err
		}
		pf.Driver = driverInfo.Name
		pf.DriverVersion = driverInfo.DriverVersion
		pf.FwVersion = driverInfo.FirmwareVersion
		pf.Vfs = Vfs{}
		err = pf.GetVfsConfig()
		if err != nil {
			return pf, err
		}
	} else {
		return pf, errors.New(fmt.Sprintf("pci device not an ethernet PF: %s", address))
	}
	return pf, nil
}

func IsPf(address string) bool {
	dir := fmt.Sprintf(checkIfPf, address)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return true
	}
	return false
}

func (pf *Pf) Stats() (map[string]uint64, error) {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	defer ethHandle.Close()
	return ethHandle.Stats(pf.InterfaceName)
}

func (pf *Pf) Features() (map[string]bool, error) {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	defer ethHandle.Close()
	return ethHandle.Features(pf.InterfaceName)
}
func (pf *Pf) LinkState() (uint32, error) {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		return 0, err
	}
	defer ethHandle.Close()
	return ethHandle.LinkState(pf.InterfaceName)
}

func (pf *Pf) GetVfsConfig() error {
	if pf.Address == "" {
		return errors.New("no pf pci address available")
	}
	totalVfsFilename := fmt.Sprintf(pfTotalVfsPath, pf.Address)
	// if the pf is not vf capable, do not try to get config
	if !DoesFileExist(totalVfsFilename) {
		return nil
	}
	totalVfs, err := ReadFileInt(totalVfsFilename)
	if err != nil {
		return err
	}
	pf.TotalVfs = totalVfs

	numVfsFilename := fmt.Sprintf(pfNumVfsPath, pf.Address)
	numVfs, err := ReadFileInt(numVfsFilename)
	if err != nil {
		return err
	}
	pf.NumVfs = numVfs

	return nil
}

// apply the pf-policy concrete config
// we don't want to return if there is
// a single error - return slice of errors
func ApplyPfConfigs(pfConfigs *[]PfConfig) []error {
	errs := make([]error, 0)
	for _, pfConfig := range *pfConfigs {
		err := SetNumVfs(pfConfig.Address, pfConfig.NumVfs)
		if err != nil {
			errs = append(errs, err)
		}
		if pfConfig.Vfio {
			pf, err := GetPfByAddress(pfConfig.Address)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, vf := range pf.Vfs {
				err := vf.BindVfio()
				if err != nil {
					errs = append(errs, err)
				}
			}
		} else {
		}
	}
	return errs
}

func SetNumVfs(address string, num int) error {
	file := fmt.Sprintf(pfNumVfsPath, address)
	err := ioutil.WriteFile(file, []byte(strconv.Itoa(num)), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (pf *Pf) SetNumVfs(num int) error {
	return SetNumVfs(pf.Address, num)
}

func (pf *Pf) PopulateLinkAttrs() error {
	link, err := netlink.LinkByName(pf.InterfaceName)
	if err != nil {
		return err
	}
	pf.MacAddress = link.Attrs().HardwareAddr.String()
	addresses, err := netlink.AddrList(link, 0)
	if err != nil {
		return err
	}
	for _, addr := range addresses {
		pf.IPAddresses = append(pf.IPAddresses, addr.IPNet.String())

	}
	if link.Attrs().Slave != nil {
		switch l := link.Attrs().Slave.(type) {
		case *netlink.BondSlave:
			pf.BondMember = l.AggregatorId
		default:
			pf.BondMember = 0
		}
	} else {
		pf.BondMember = 0
	}
	// search for sub interfaces (dot1q)
	links, err := netlink.LinkList()
	if err != nil {
		return nil
	}
	for _, l := range links {
		if l.Attrs().ParentIndex == link.Attrs().Index {
			pf.HasSubinterfaces = true
			addresses, err := netlink.AddrList(link, 0)
			if err != nil {
				return err
			}
			for _, addr := range addresses {
				pf.IPAddresses = append(pf.IPAddresses, addr.IPNet.String())

			}
		}
	}
	if len(pf.IPAddresses) > 0 {
		pf.HasIPAddress = true
	} else {
		pf.HasIPAddress = false
	}
	return nil
}

func (pf *Pf) IPAddressesJoin() string {
	return strings.Join(pf.IPAddresses, ",")
}

// handle (socket) for the network requests on a specific network namespace.
func (pf *Pf) SetVfMacAddresses() []error {
	errs := make([]error, 0)
	link, err := netlink.LinkByName(pf.InterfaceName)
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	for n, _ := range pf.Vfs {
		hwAddr := GenerateMac()
		err := netlink.LinkSetVfHardwareAddr(link, n, hwAddr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errs
}

func (pf *Pf) SetVfMacAddressesNetlink() error {
	_, err := netlink.LinkByName(pf.InterfaceName)
	if err != nil {
		return err
	}
	return nil
}

func PfFromVfAddress(address string) Pf {
	interfaceName := InterfaceNameFromAddress(address)
	if interfaceName == "" {
		return Pf{}
	}
	driverInfo, err := KernelDeviceDriver(interfaceName)
	if err != nil {
		return Pf{}
	}
	return Pf{
		InterfaceName: interfaceName,
		Address:       address,
		Driver:        driverInfo.Name,
		DriverVersion: driverInfo.DriverVersion,
		FwVersion:     driverInfo.FirmwareVersion,
	}
}
