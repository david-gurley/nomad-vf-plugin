package host

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jaypipes/ghw"
)

const (
	hostIdPath = "/etc/machine-id"
)

func GetHost() (*Host, error) {
	var host Host
	host.OS = runtime.GOOS
	host.CPUs = runtime.NumCPU()
	hostname, err := GetHostname()
	if err != nil {
		return &host, err
	}
	host.Hostname = hostname
	hostID, err := ReadFileString(hostIdPath)
	if err != nil {
		return &host, err
	}
	host.HostID = hostID
	err = host.updateUname()
	if err != nil {
		return &host, err
	}
	err = host.updateLsbRelease()
	if err != nil {
		return &host, err
	}
	cpu, err := ghw.CPU()
	if err != nil {
		return &host, err
	}
	host.CPUVendor = cpu.Processors[0].Vendor
	host.CPUModel = cpu.Processors[0].Model
	pfs, err := GetPfs()
	if err != nil {
		return &host, err
	}
	host.Pfs = pfs
	return &host, nil
}

func (host *Host) updateUname() error {
	cmd := exec.Command("uname", "-srio")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	o := out.String()
	osStr := strings.Replace(o, "\n", "", -1)
	osStr = strings.Replace(osStr, "\r\n", "", -1)
	osInfo := strings.Split(osStr, " ")
	host.Kernel = osInfo[0]
	host.Core = osInfo[1]
	host.Platform = osInfo[2]
	return nil
}

func (host *Host) updateLsbRelease() error {
	cmd := exec.Command("lsb_release", "-a")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	o := out.String()
	lsbInfo := strings.Split(o, "\n")

	host.Distribution = strings.Split(lsbInfo[0], "\t")[1]
	host.DistributorID = strings.Split(lsbInfo[1], "\t")[1]
	host.Release = strings.Split(lsbInfo[2], "\t")[1]
	host.Codename = strings.Split(lsbInfo[3], "\t")[1]
	return nil
}

func GetHostID() (string, error) {
	return ReadFileString(hostIdPath)
}

// cat /proc/cmdline
// BOOT_IMAGE=/vmlinuz-5.4.0-81-generic root=/dev/mapper/ubuntu--vg-ubuntu--lv ro intel_iommu=on iommu=pt vfio_iommu_type1.allow_unsafe_interrupts=1
func GetKernelCmdline() (string, error) {
	return "", nil
}
