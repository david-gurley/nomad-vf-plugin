package vf

import (
	"context"
	"time"

	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"

	"github.com/david-gurley/host"
)

const (
	// attributes for device groups (pf-level)
	PfDriverAttr          = "pf_driver"
	PfDriverVersionAttr   = "pf_driver_version"
	PfFirmwareVersionAttr = "pf_firmware_version"
)

// doFingerprint is the long-running goroutine that detects device changes
func (d *VfDevicePlugin) doFingerprint(ctx context.Context, devices chan *device.FingerprintResponse) {
	defer close(devices)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(d.fingerprintPeriod)
		}

		d.writeFingerprintToChannel(devices)
	}
}

// build fingerprint/stats response with computed groups
// {{ vendor }}/{{ device_type }}/{{ pf_address }}
// e.g. pensando/vf/0000.0000.0000
type GroupMapping struct {
	Devices host.Vfs
	Vendor  string
	Type    string
}

// writeFingerprintToChannel collects fingerprint info, partitions devices into
// device groups, and sends the data over the provided channel.
func (d *VfDevicePlugin) writeFingerprintToChannel(devices chan<- *device.FingerprintResponse) {

	fingerprintData, err := host.GetVfs()
	if err != nil {
		d.logger.Error("failed to get fingerprint pci vf devices", "error", err)
		devices <- device.NewFingerprintError(err)
		return
	}
	pfsMap, err := host.GetPfsMap()
	if err != nil {
		d.logger.Error("failed to get fingerprint pci pf devices", "error", err)
		devices <- device.NewFingerprintError(err)
		return
	}

	// only show devices we care about (from configuration)
	fingerprintDevices := filterFingerprintedDevices(fingerprintData, d.vendors)

	deviceGroupNames := make(map[string]GroupMapping)
	devicesMap := make(map[string]*host.Vf)
	for _, vf := range fingerprintDevices {
		if vf.Allocated {
			continue
		}
		devicesMap[vf.Address] = vf
		deviceGroupNames[vf.PfAddress] = GroupMapping{
			Devices: append(deviceGroupNames[vf.PfAddress].Devices, vf),
			Vendor:  vf.Vendor,
			Type:    "vf",
		}
	}
	d.devices = devicesMap
	numDeviceGroups := 0
	for _, _ = range deviceGroupNames {
		numDeviceGroups++
	}
	deviceGroups := make([]*device.DeviceGroup, 0, numDeviceGroups)
	for groupName, groupMapping := range deviceGroupNames {
		devices := make([]*device.Device, 0)
		for _, vf := range groupMapping.Devices {
			devices = append(devices, &device.Device{
				ID:      vf.Address,
				Healthy: true,
				HwLocality: &device.DeviceLocality{
					PciBusID: vf.Address,
				},
			})
		}
		deviceGroups = append(deviceGroups, &device.DeviceGroup{
			Vendor:     groupMapping.Vendor,
			Type:       groupMapping.Type,
			Name:       groupName,
			Devices:    devices,
			Attributes: attributesFromFingerprintDeviceData(groupMapping.Devices, pfsMap[groupMapping.Devices[0].PfAddress]),
		})
	}
	devices <- device.NewFingerprint(deviceGroups...)
}

// ignoreFingerprintedDevices excludes ignored devices from fingerprint output
func filterFingerprintedDevices(deviceData host.Vfs, vendors []string) host.Vfs {
	return deviceData.ByVendors(vendors)
}

// attributes for a slice of vfs associated to a single pf
func attributesFromFingerprintDeviceData(d host.Vfs, pf *host.Pf) map[string]*structs.Attribute {
	attrs := map[string]*structs.Attribute{}
	attrs[PfFirmwareVersionAttr] = &structs.Attribute{String: &pf.FwVersion}
	attrs[PfDriverAttr] = &structs.Attribute{String: &pf.Driver}
	attrs[PfDriverVersionAttr] = &structs.Attribute{String: &pf.DriverVersion}
	return attrs
}
