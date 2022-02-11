package vf

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/kr/pretty"

	"github.com/david-gurley/host"
)

const (
	pluginName    = "vf"
	pluginVersion = "v0.1.0"
	vendor        = "generic"
	deviceType    = "vfio-pci"
)

var (
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDevice,
		PluginApiVersions: []string{device.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"enabled": hclspec.NewDefault(
			hclspec.NewAttr("enabled", "bool", false),
			hclspec.NewLiteral("true"),
		),
		"vendors": hclspec.NewDefault(
			hclspec.NewAttr("vendors", "list(string)", false),
			hclspec.NewLiteral("[\"pensando\"]"),
		),
		"fingerprint_period": hclspec.NewDefault(
			hclspec.NewAttr("fingerprint_period", "string", false),
			hclspec.NewLiteral("\"1m\""),
		),
	})
)

type Config struct {
	Enabled           bool     `codec:"enabled"`
	Vendors           []string `codec:"vendors"`
	FingerprintPeriod string   `codec:"fingerprint_period"`
}

type VfDevicePlugin struct {
	logger            log.Logger
	enabled           bool
	vendors           []string
	fingerprintPeriod time.Duration
	devices           map[string]*host.Vf
	deviceLock        sync.RWMutex
}

// initialize any map or slice attributes
func NewPlugin(log log.Logger) *VfDevicePlugin {
	return &VfDevicePlugin{
		logger:  log.Named(pluginName),
		devices: make(map[string]*host.Vf),
		vendors: make([]string, 1),
	}
}

func (d *VfDevicePlugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

func (d *VfDevicePlugin) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

func (d *VfDevicePlugin) SetConfig(c *base.Config) error {
	var config Config
	if err := base.MsgPackDecode(c.PluginConfig, &config); err != nil {
		return err
	}
	d.enabled = config.Enabled
	d.vendors = config.Vendors

	period, err := time.ParseDuration(config.FingerprintPeriod)
	if err != nil {
		return fmt.Errorf("failed to parse doFingerprint period %q: %v", config.FingerprintPeriod, err)
	}
	d.fingerprintPeriod = period
	d.logger.Info("config set", "config", log.Fmt("% #v", pretty.Formatter(config)))
	return nil
}

func (d *VfDevicePlugin) Fingerprint(ctx context.Context) (<-chan *device.FingerprintResponse, error) {
	outCh := make(chan *device.FingerprintResponse)
	go d.doFingerprint(ctx, outCh)
	return outCh, nil
}

func (d *VfDevicePlugin) Stats(ctx context.Context, interval time.Duration) (<-chan *device.StatsResponse, error) {
	outCh := make(chan *device.StatsResponse)
	go d.doStats(ctx, outCh, interval)
	return outCh, nil
}

type reservationError struct {
	notExistingIDs []string
}

func (e *reservationError) Error() string {
	return fmt.Sprintf("unknown device IDs: %s", strings.Join(e.notExistingIDs, ","))
}

func (d *VfDevicePlugin) Reserve(deviceIDs []string) (*device.ContainerReservation, error) {
	if len(deviceIDs) == 0 {
		return &device.ContainerReservation{}, nil
	}

	d.deviceLock.RLock()
	var notExistingIDs []string
	for _, deviceId := range deviceIDs {
		if _, deviceIDExists := d.devices[deviceId]; !deviceIDExists {
			notExistingIDs = append(notExistingIDs, deviceId)
		}
	}

	d.deviceLock.RUnlock()
	if len(notExistingIDs) != 0 {
		return nil, &reservationError{notExistingIDs}
	}

	envs := make(map[string]string)
	for i, id := range deviceIDs {
		envs[fmt.Sprintf("DEVICE_VF_%s_%d", d.devices[id].Vendor, i)] = id

	}

	return &device.ContainerReservation{
		Envs: envs,
	}, nil
}
