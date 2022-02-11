package vf

import (
	"context"
	"time"

	"github.com/david-gurley/host"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

// doStats is the long running goroutine that streams device statistics
func (d *VfDevicePlugin) doStats(ctx context.Context, stats chan<- *device.StatsResponse, interval time.Duration) {
	defer close(stats)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(interval)
		}

		d.writeStatsToChannel(stats, time.Now())
	}
}

// deviceStats is what we "collect" and transform into device.DeviceStats objects.
//
// could get stats on the PF from ethtool
type deviceStats struct {
	Address   string
	Allocated bool
}

// writeStatsToChannel collects device stats, partitions devices into
// device groups, and sends the data over the provided channel.
func (d *VfDevicePlugin) writeStatsToChannel(stats chan<- *device.StatsResponse, timestamp time.Time) {

	pfsMap, err := host.GetPfsMap()
	if err != nil {
		d.logger.Error("error getting pfs map", "error", err)
	}
	deviceGroupNames := make(map[string]GroupMapping)
	for _, vf := range d.devices {
		deviceGroupNames[vf.PfAddress] = GroupMapping{
			Devices: append(deviceGroupNames[vf.PfAddress].Devices, vf),
			Vendor:  vf.Vendor,
			Type:    "vf",
		}
	}
	deviceGroupStats := make([]*device.DeviceGroupStats, 0)
	for groupName, groupMapping := range deviceGroupNames {
		pfStats, err := pfsMap[groupName].Stats()
		if err != nil {
			continue
		}
		txBytes := pfStats["tx_bytes"]
		rxBytes := pfStats["rx_byptes"]
		deviceStats := &device.DeviceStats{
			Summary: &structs.StatValue{
				Desc:            "Tx Bytes",
				IntNumeratorVal: uint64ToInt64Ptr(&txBytes),
				Unit:            "Bytes",
			},
			Stats: &structs.StatObject{
				Attributes: map[string]*structs.StatValue{
					"tx_bytes": &structs.StatValue{
						Desc:            "Tx Bytes",
						IntNumeratorVal: uint64ToInt64Ptr(&txBytes),
						Unit:            "Bytes",
					},
					"rx_bytes": &structs.StatValue{
						Desc:            "Rx Bytes",
						IntNumeratorVal: uint64ToInt64Ptr(&rxBytes),
						Unit:            "Bytes",
					},
				},
			},
			Timestamp: time.Now(),
		}
		instanceStats := make(map[string]*device.DeviceStats)
		for _, vf := range groupMapping.Devices {
			instanceStats[vf.Address] = deviceStats
		}
		deviceGroupStats = append(deviceGroupStats, &device.DeviceGroupStats{
			Vendor:        groupMapping.Vendor,
			Type:          groupMapping.Type,
			Name:          groupName,
			InstanceStats: instanceStats,
		})
	}

	stats <- &device.StatsResponse{
		Groups: deviceGroupStats,
	}
}

func uintToInt64Ptr(u *uint) *int64 {
	if u == nil {
		return nil
	}

	v := int64(*u)
	return &v
}

func uint64ToInt64Ptr(u *uint64) *int64 {
	if u == nil {
		return nil
	}

	v := int64(*u)
	return &v
}
