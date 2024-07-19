package netplan

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

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

func (nc *NetplanConfig) Write() ([]byte, error) {
	bytes, err := yaml.Marshal(nc)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot write config file. error: %s", err)
	}
	return bytes, nil
}

func (nc *NetplanConfig) WriteFile(configPath string) error {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	bytes, err := nc.Write()
	if err != nil {
		return fmt.Errorf("cannot write config file. error: %s", err)
	}

	err = os.WriteFile(configPath, bytes, os.FileMode(int(0644)))
	if err != nil {
		return fmt.Errorf("cannot write config file. error: %s", err)
	}

	return nil
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
		return fmt.Errorf("interface not found: %s", name)
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

func (nc *NetplanConfig) SetDhcp4(name string, enable bool) error {
	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		return fmt.Errorf("interface not found: %s", name)
	}

	// update ips
	objs[name].SetDhcp4(enable)

	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}

func (nc *NetplanConfig) SetGateway4(name string, ip IP) error {
	if ip.IsCIDR() {
		return fmt.Errorf("invalid IP (%s), it should not be CIDR", ip)
	}

	objs := nc.Flatten()
	if _, ok := objs[name]; !ok {
		return fmt.Errorf("interface not found: %s", name)
	}

	// update ips
	objs[name].SetGateway4(ip)

	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}

func (nc *NetplanConfig) AddBond(name, confInterface string, interfaces []string) error {
	if name == "" || confInterface == "" {
		return errors.New("invalid interface")
	}
	if len(interfaces) == 0 {
		return errors.New("empty interface(s)")
	}
	if !slices.Contains(interfaces, confInterface) {
		return errors.New("configuration interface must be included within the interfaces")
	}

	objs := nc.Flatten()
	var chkIfaces []string
	copy(chkIfaces, interfaces)
	for n := range objs {
		if idx := slices.Index(chkIfaces, n); idx != -1 {
			chkIfaces = append(chkIfaces[:idx], chkIfaces[idx+1:]...)
		}
	}
	if len(chkIfaces) > 0 {
		return fmt.Errorf("interface(s) not found: %s", strings.Join(chkIfaces, ", "))
	}

	if _, ok := objs[confInterface]; !ok {
		return fmt.Errorf("config interface not found: %s", confInterface)
	}

	b := &Bond{}
	b.Interfaces = interfaces
	b.Copy(objs[confInterface])
	objs[name] = b
	objs[confInterface] = &Ethernet{}

	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}

func (nc *NetplanConfig) GetBond(name string) (*Bond, error) {
	for n, o := range nc.Network.Bonds {
		if n == name {
			return &o, nil
		}
	}

	return nil, fmt.Errorf("bond not found:%s", name)
}

func (nc *NetplanConfig) RemoveBond(name, confInterface string) error {
	objs := nc.Flatten()

	if _, ok := objs[name]; !ok {
		return fmt.Errorf("bonding not found: %s", name)
	}

	var b Bond
	switch v := objs[name].(type) {
	case *Bond:
		b = *v
	default:
		return fmt.Errorf("interface is not bonding: %s", name)
	}

	objs[confInterface] = b.Dump()
	delete(objs, name)
	newConfig, err := GetConfig(objs)
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	*nc = *newConfig

	return nil
}
