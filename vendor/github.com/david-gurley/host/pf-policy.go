package host

import (
	"errors"

	"github.com/jinzhu/copier"
)

const (
	PF_POLICY_CURRENT = "current"
	PF_POLICY_PLANNED = "planned"
)

var (
	ErrPlanRequired = errors.New("pf-plan required by not available")
)

type PfPolicyOptions struct {
	OnReboot bool `json:"on_reboot"`
	MaxVfs   bool `json:"max_vfs"`
	NumVfs   int  `json:"num_vfs"`
}

// the host does not need to know id or version
type PfPolicy struct {
	PfSelector *PfSelector `json:"pf_selector"`
	OnReboot   bool        `json:"on_reboot"`
	MaxVfs     bool        `json:"max_vfs"`
	NumVfs     int         `json:"num_vfs"`
	// result of a state change to PLAN
	// host_id: pf_id: before|after
	PfPlan map[string]map[string]map[string]*Pf `json:"pf_plan"`
}

func NewPfPolicy(pfSelectorOpts *PfSelectorOptions, pfPolicyOpts *PfPolicyOptions) (*PfPolicy, error) {
	pfSelector, err := NewPfSelector(pfSelectorOpts)
	if err != nil {
		return &PfPolicy{}, err
	}
	pfPolicy := &PfPolicy{
		PfSelector: pfSelector,
		OnReboot:   pfPolicyOpts.OnReboot,
		MaxVfs:     pfPolicyOpts.MaxVfs,
		NumVfs:     pfPolicyOpts.NumVfs,
	}
	pfPolicy.PfPlan = make(map[string]map[string]map[string]*Pf)
	return pfPolicy, nil
}
func (pfPolicy *PfPolicy) Plan() error {
	for _, pfs := range *pfPolicy.PfSelector.Selected {
		for _, pf := range pfs {
			if pfPolicy.PfPlan[pf.HostID] == nil {
				pfPolicy.PfPlan[pf.HostID] = make(map[string]map[string]*Pf)
			}
			if pfPolicy.PfPlan[pf.HostID][pf.Address] == nil {
				pfPolicy.PfPlan[pf.HostID][pf.Address] = make(map[string]*Pf)
			}
			pfPolicy.PfPlan[pf.HostID][pf.Address][PF_POLICY_CURRENT] = &pf
			pfPlanned := &Pf{}
			copier.Copy(pfPlanned, pf)
			if pf.TotalVfs > 0 && pfPolicy.MaxVfs {
				pfPlanned.NumVfs = pf.TotalVfs
			} else if pf.NumVfs > 0 && pfPolicy.NumVfs > 0 {
				pfPlanned.NumVfs = pfPolicy.NumVfs

			} else {
				// we get here if the pf is not vf capable
			}
			pfPolicy.PfPlan[pf.HostID][pf.Address][PF_POLICY_PLANNED] = pfPlanned
		}
	}
	return nil
}

func (pfPolicy *PfPolicy) Apply(hostID string) error {
	if pfPolicy.PfPlan == nil {
		return ErrPlanRequired
	}
	if pfPolicy.PfPlan[hostID] == nil {
		return ErrPlanRequired
	}
	for _, plan := range pfPolicy.PfPlan[hostID] {
		// num_vfs
		if plan[PF_POLICY_PLANNED].NumVfs == plan[PF_POLICY_CURRENT].NumVfs {
			continue
		}
		err := plan[PF_POLICY_PLANNED].SetNumVfs(plan[PF_POLICY_PLANNED].NumVfs)
		if err != nil {
			return err
		}
	}
	return nil
}

// agent uses to calculate concrete from logical provided from server
func (pfPolicy *PfPolicy) ApplyConcrete(hostID string) (*[]PfConfig, error) {
	pfConfigs := make([]PfConfig, 0)
	if pfPolicy.PfPlan == nil {
		return &pfConfigs, ErrPlanRequired
	}
	if pfPolicy.PfPlan[hostID] == nil {
		return &pfConfigs, ErrPlanRequired
	}
	for address, plan := range pfPolicy.PfPlan[hostID] {
		pfConfigs = append(pfConfigs, PfConfig{
			Address: address,
			NumVfs:  plan[PF_POLICY_PLANNED].NumVfs,
			Vfio:    true,
		})
	}
	return &pfConfigs, nil
}
