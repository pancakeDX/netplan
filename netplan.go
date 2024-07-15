package netplan

import (
	"errors"
	"fmt"
	"os"

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

func GetConfig(objs map[string]Layout) (*NetplanConfig, error) {
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
		case *Bond:
			n.Network.Bonds[name] = *v
		case *Ethernet:
			n.Network.Ethernets[name] = *v
		case *Vlan:
			n.Network.Vlans[name] = *v
		default:
			return nil, fmt.Errorf("invalid type: %s", v)
		}
	}

	return n, nil
}

func (nc *NetplanConfig) Flatten() map[string]Layout {
	objs := make(map[string]Layout)

	// Ethernets
	for n, o := range nc.Network.Ethernets {
		objs[n] = &o
	}

	// Bonding
	for n, o := range nc.Network.Bonds {
		objs[n] = &o
	}

	// Vlan
	for n, o := range nc.Network.Vlans {
		objs[n] = &o
	}

	return objs
}

func (nc *NetplanConfig) GetAddr(name string, ips []IP) ([]IP, error) {
	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		return nil, fmt.Errorf("interface not found: %s", name)
	}

	return objs[name].GetAddrs(), nil
}

func (nc *NetplanConfig) SetAddr(name string, ips []IP) error {
	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		return fmt.Errorf("interface not found: %s", name)
	}

	// update ips
	objs[name].UpdateAddrs(ips)

	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}
