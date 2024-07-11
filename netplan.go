package netplan

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigVersion = 2
	DefaultRenderer      = "networkd"
)

// NetplanConfig represents the top-level netplan configuration with network layer
type NetplanConfig struct {
	Network Netplan `yaml:"network"`
}

// Netplan represents the netplan configuration inside the network layer
type Netplan struct {
	Version   int                 `yaml:"version,omitempty"`
	Renderer  string              `yaml:"renderer,omitempty"`
	Ethernets map[string]Ethernet `yaml:"ethernets,omitempty"`
	Bonds     map[string]Bond     `yaml:"bonds,omitempty"`
	Vlans     map[string]Vlan     `yaml:"vlans,omitempty"`
}

// Common represents the common fields for various network configurations
type Common struct {
	Addresses   []IP        `yaml:"addresses,omitempty"`
	Gateway4    string      `yaml:"gateway4,omitempty"`
	Nameservers Nameservers `yaml:"nameservers,omitempty"`
	Dhcp4       bool        `yaml:"dhcp4,omitempty"`
	Dhcp6       bool        `yaml:"dhcp6,omitempty"`
}

// Ethernet represents an ethernet interface
type Ethernet struct {
	Common `yaml:",inline"`
}

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

// Vlan represents a vlan interface
type Vlan struct {
	Common `yaml:",inline"`
	ID     int    `yaml:"id,omitempty"`
	Link   string `yaml:"link,omitempty"`
}

// Nameservers represents the nameservers configuration for a vlan interface
type Nameservers struct {
	Addresses []IP `yaml:"addresses,omitempty"`
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

func (b *BondMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	mode, ok := stringToBondMode[str]
	if !ok {
		*b = BondModeUnknown
		return errors.New("unknown bond mode")
	}

	*b = mode
	return nil

}

func (b BondMode) MarshalYAML() (interface{}, error) {
	mode, ok := bondModeToString[b]
	if !ok {
		return "", errors.New("unknown bond mode")
	}

	return mode, nil
}

func Read(config []byte) (*NetplanConfig, error) {
	var readConfig NetplanConfig
	err := yaml.Unmarshal(config, &readConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file. error: %s", err)
	}

	return &readConfig, nil
}

func Write(config *NetplanConfig) ([]byte, error) {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot write config file. error: %s", err)
	}
	return bytes, nil
}

func GetConfig(objs map[string]interface{}) (*NetplanConfig, error) {
	n := &NetplanConfig{
		Network: Netplan{
			Version:   DefaultConfigVersion,
			Renderer:  DefaultRenderer,
			Bonds:     make(map[string]Bond),
			Ethernets: make(map[string]Ethernet),
			Vlans:     make(map[string]Vlan),
		},
	}

	for name, i := range objs {
		switch v := i.(type) {
		case Bond:
			n.Network.Bonds[name] = i.(Bond)
		case Ethernet:
			n.Network.Ethernets[name] = i.(Ethernet)
		case Vlan:
			n.Network.Vlans[name] = i.(Vlan)
		default:
			return nil, fmt.Errorf("invalid type: %s", v)
		}
	}

	return n, nil
}
