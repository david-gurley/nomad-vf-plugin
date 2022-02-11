package host

import (
	"regexp"
)

const (
	PF_SELECTOR_KIND_VENDOR_REGEXP    = "vendor_regexp"
	PF_SELECTOR_VENDOR_REGEXP_DEFAULT = ".*"
)

type PfSelectorOptions struct {
	Kind         string `json:"kind"`
	VendorRegexp string `json:"vendor_regexp"`
}

type PfSelector struct {
	Kind         string           `json:"kind"`
	VendorRegexp string           `json:"vendor_regexp"`
	Selected     *map[string][]Pf `json:"selected"` // by host_id
}

func NewPfSelector(options *PfSelectorOptions) (*PfSelector, error) {
	var pfSelector PfSelector
	switch options.Kind {
	case PF_SELECTOR_KIND_VENDOR_REGEXP:
		pfSelector.Kind = PF_SELECTOR_KIND_VENDOR_REGEXP
		_, err := regexp.Compile(options.VendorRegexp)
		if err != nil {
			return &pfSelector, err
		}
		pfSelector.VendorRegexp = options.VendorRegexp
	default:
		pfSelector.Kind = PF_SELECTOR_KIND_VENDOR_REGEXP
		pfSelector.VendorRegexp = PF_SELECTOR_VENDOR_REGEXP_DEFAULT
		_, err := regexp.Compile(pfSelector.VendorRegexp)
		if err != nil {
			return &pfSelector, err
		}
	}
	return &pfSelector, nil
}

// this is used by host server to select from a slice of hosts that have
// already been selected by the host selector
func (pfSelector *PfSelector) Select(hosts *[]Host) error {
	selectedPfs := make(map[string][]Pf)
	switch pfSelector.Kind {
	case PF_SELECTOR_KIND_VENDOR_REGEXP:
		r, err := regexp.Compile(pfSelector.VendorRegexp)
		if err != nil {
			return err
		}
		for _, h := range *hosts {
			for _, pf := range h.Pfs {
				if r.MatchString(pf.Vendor) {
					selectedPfs[h.HostID] = append(selectedPfs[h.HostID], *pf)
				}
			}
		}
		pfSelector.Selected = &selectedPfs
	default:
	}
	return nil
}

// this is used by the agent locally to select pfs based on the selector configuration
func (pfSelector *PfSelector) SelectLocalhost() error {
	host, err := GetHost()
	if err != nil {
		return err
	}
	selectedPfs := make(map[string][]Pf)
	switch pfSelector.Kind {
	case PF_SELECTOR_KIND_VENDOR_REGEXP:
		r, err := regexp.Compile(pfSelector.VendorRegexp)
		if err != nil {
			return err
		}
		for _, pf := range host.Pfs {
			if r.MatchString(pf.Vendor) {
				if selectedPfs[host.HostID] == nil {
					selectedPfs[host.HostID] = []Pf{}
				}
				selectedPfs[host.HostID] = append(selectedPfs[host.HostID], *pf)
			}
		}
		pfSelector.Selected = &selectedPfs
	}
	return nil
}
