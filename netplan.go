package netplan

import (
	"errors"
	"fmt"
	gnet "net"
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

func ReadFile(configPath string) (*NetplanConfig, error) {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	config, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file. error: %s", err)
	}

	nc, err := Read(config)
	if err != nil {
		return nil, err
	}

	return nc, nil
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
		newO := o
		objs[n] = &newO
	}

	// Bonding
	for n, o := range nc.Network.Bonds {
		newO := o
		objs[n] = &newO
	}

	// Vlan
	for n, o := range nc.Network.Vlans {
		newO := o
		objs[n] = &newO
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
	for _, ip := range ips {
		if !ip.IsCIDR() {
			return fmt.Errorf("invalid IP (%s), it should be CIDR", ip)
		}
	}

	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		ifaces, err := gnet.Interfaces()
		if err != nil {
			return fmt.Errorf("error getting interface: %s", err)
		}
		chk := false
		var ifType InterfaceType
		for _, iface := range ifaces {
			if iface.Name == name {
				ifType, _ = getInterfaceType(iface.Name)
				chk = true
			}
		}
		if !chk {
			return fmt.Errorf("interface not found: %s", name)
		}

		if ifType == InterfaceTypeEthernet {
			objs[name] = &Ethernet{}
		} else if ifType == InterfaceTypeBonding {
			objs[name] = &Bond{}
		} else if ifType == InterfaceTypeVLAN {
			objs[name] = &Vlan{}
		} else {
			return fmt.Errorf("unsupported interface type: %s", name)
		}
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

func (nc *NetplanConfig) GetNS(name string, ips []IP) ([]IP, error) {
	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		return nil, fmt.Errorf("interface not found: %s", name)
	}

	return objs[name].GetNS(), nil
}

func (nc *NetplanConfig) SetNS(name string, ips []IP) error {
	for _, ip := range ips {
		if ip.IsCIDR() {
			return fmt.Errorf("invalid IP (%s), it should not be CIDR", ip)
		}
	}

	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		ifaces, err := gnet.Interfaces()
		if err != nil {
			return fmt.Errorf("error getting interface: %s", err)
		}
		chk := false
		var ifType InterfaceType
		for _, iface := range ifaces {
			if iface.Name == name {
				ifType, _ = getInterfaceType(iface.Name)
				chk = true
			}
		}
		if !chk {
			return fmt.Errorf("interface not found: %s", name)
		}

		if ifType == InterfaceTypeEthernet {
			objs[name] = &Ethernet{}
		} else if ifType == InterfaceTypeBonding {
			objs[name] = &Bond{}
		} else if ifType == InterfaceTypeVLAN {
			objs[name] = &Vlan{}
		} else {
			return fmt.Errorf("unsupported interface type: %s", name)
		}
	}

	// update ips
	objs[name].UpdateNS(ips)

	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}
