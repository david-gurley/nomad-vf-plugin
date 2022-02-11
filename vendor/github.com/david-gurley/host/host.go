package host

import (
	"errors"

	"github.com/Showmax/go-fqdn"
)

type Host struct {
	HostID        string `json:"host_id"`
	Hostname      string `json:"hostname"`
	OS            string `json:"os"`
	Kernel        string `json:"kernel"`
	Core          string `json:"core"`
	Platform      string `json:"platform"`
	CPUs          int    `json:"cpus"`
	CPUVendor     string `json:"cpu_vendor"`
	CPUModel      string `json:"cpu_model"`
	Distribution  string `json:"distribution"`
	DistributorID string `json:"distributor_id"`
	Release       string `json:"release"`
	Codename      string `json:"codename"`
	Pfs           Pfs    `json:"pfs"`
}

func GetHostname() (string, error) {
	name, err := fqdn.FqdnHostname()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("got an empty fqdn")
	}
	return name, nil
}
