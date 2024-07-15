package netplan

import "errors"

// Bond represents a bond interface
type Bond struct {
	Common     `yaml:",inline"`
	Interfaces []string       `yaml:"interfaces,omitempty"`
	Parameters BondParameters `yaml:"parameters,omitempty"`
}

// BondParameters represents the parameters for a bond interface
type BondParameters struct {
	Mode               BondMode `yaml:"mode,omitempty"`
	MiiMonitorInterval int      `yaml:"mii-monitor-interval,omitempty"`
	TransmitHashPolicy string   `yaml:"transmit-hash-policy,omitempty"`
}

type BondMode int

const BondModeUnknown BondMode = -1

const (
	BondModeBalanceRR BondMode = iota
	BondModeActiveBackup
	BondModeBalanceXor
	BondModeBroadcast
	BondMode8023ad
	BondModeBalanceTlb
	BondModeBalanceAlb
)

var bondModeToString = map[BondMode]string{
	BondModeBalanceRR:    "balance-rr",
	BondModeActiveBackup: "active-backup",
	BondModeBalanceXor:   "balance-xor",
	BondModeBroadcast:    "broadcast",
	BondMode8023ad:       "802.3ad",
	BondModeBalanceTlb:   "balance-tlb",
	BondModeBalanceAlb:   "balance-alb",
}

var stringToBondMode = map[string]BondMode{
	"balance-rr":    BondModeBalanceRR,
	"active-backup": BondModeActiveBackup,
	"balance-xor":   BondModeBalanceXor,
	"broadcast":     BondModeBroadcast,
	"802.3ad":       BondMode8023ad,
	"balance-tlb":   BondModeBalanceTlb,
	"balance-alb":   BondModeBalanceAlb,
}

func (bm *BondMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	mode, ok := stringToBondMode[str]
	if !ok {
		*bm = BondModeUnknown
		return errors.New("unknown bond mode")
	}

	*bm = mode
	return nil

}

func (bm BondMode) MarshalYAML() (interface{}, error) {
	mode, ok := bondModeToString[bm]
	if !ok {
		return "", errors.New("unknown bond mode")
	}

	return mode, nil
}

func (b *Bond) UpdateAddrs(ips []IP) {
	b.Addresses = ips
}

func (b *Bond) GetAddrs() []IP {
	return b.Addresses
}

func (b *Bond) UpdateNS(ips []IP) {
	b.Nameservers.Addresses = ips
}

func (b *Bond) GetNS() []IP {
	return b.Nameservers.Addresses
}

func (b *Bond) SetDhcp4(enable bool) {
	b.Dhcp4 = enable
}

func (b *Bond) SetGateway4(ip IP) {
	b.Gateway4 = ip
}
