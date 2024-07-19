package netplan

import (
	gnet "net"
	"os"
	"reflect"
	"strings"
	"testing"

	"gotest.tools/assert"
)

var (
	config = `network:
  version: 2
  renderer: networkd
  ethernets:
    eth0: {}
    eth1: {}
  bonds:
    bond0:
      interfaces:
      - eth0
      - eth1
      parameters:
        mode: 802.3ad
  vlans:
    vlan100:
      nameservers:
        addresses:
        - 8.8.8.8
        - 8.8.4.4
      id: 100
      link: bond0
`
	netplanConfig = NetplanConfig{
		Network: Netplan{
			Version:  2,
			Renderer: "networkd",
			Ethernets: map[string]Ethernet{
				"eth0": {Common{Dhcp4: false, Dhcp6: false}},
				"eth1": {Common{Dhcp4: false, Dhcp6: false}},
			},
			Bonds: map[string]Bond{
				"bond0": {
					Interfaces: []string{"eth0", "eth1"},
					Parameters: BondParameters{Mode: BondMode8023ad},
				},
			},
			Vlans: map[string]Vlan{
				"vlan100": {
					Link: "bond0",
					ID:   100,
					Common: Common{
						Nameservers: Nameservers{
							Addresses: []IP{
								{IP: gnet.ParseIP("8.8.8.8")},
								{IP: gnet.ParseIP("8.8.4.4")},
							},
						},
					},
				},
			},
		},
	}
)

func TestRead(t *testing.T) {
	readConfig, err := Read([]byte(config))
	if err != nil {
		t.Errorf("cannot read config. error: %v", err)
	}
	eq := reflect.DeepEqual(*readConfig, netplanConfig)
	if !eq {
		t.Error("invalid config")
	}
}

func TestWrite(t *testing.T) {
	bytes, err := netplanConfig.Write()
	if err != nil {
		t.Errorf("cannot write config. error: %v", err)
	}
	assert.Equal(t, string(bytes), config)
}

func TestGetConfig(t *testing.T) {
	c := map[string]Layout{}
	eth0 := Ethernet{}
	eth1 := Ethernet{}
	ns := Nameservers{
		Addresses: []IP{
			{IP: gnet.ParseIP("8.8.8.8")},
			{IP: gnet.ParseIP("8.8.4.4")},
		},
	}
	bond0 := Bond{
		Interfaces: []string{"eth0", "eth1"},
		Parameters: BondParameters{Mode: BondMode8023ad},
	}
	vlan100 := Vlan{
		ID:   100,
		Link: "bond0",
	}
	vlan100.Nameservers = ns

	// append all to map
	c["eth0"] = &eth0
	c["eth1"] = &eth1
	c["bond0"] = &bond0
	c["vlan100"] = &vlan100

	nc, err := GetConfig(c)
	if err != nil {
		t.Errorf("cannot get config. error: %s", err)
	}

	eq := reflect.DeepEqual(*nc, netplanConfig)
	if !eq {
		t.Error("invalid config")
	}
}

func TestFlatten(t *testing.T) {
	c := map[string]Layout{}
	eth0 := Ethernet{}
	ip, _ := ParseIP("1.2.3.4")
	eth0.Addresses = []IP{*ip}
	c["eth0"] = &eth0

	nc, err := GetConfig(c)
	if err != nil {
		t.Errorf("cannot get config. error: %s", err)
	}

	ncFlat := nc.Flatten()
	if _, ok := ncFlat["eth0"]; !ok {
		t.Errorf("cannot get config from flat config. error: %s", err)
	}

	chkIP, _ := ParseIP("1.2.3.4")
	chk := false
	for _, addr := range ncFlat["eth0"].GetAddrs() {
		if addr.String() == chkIP.String() {
			chk = true
		}
	}

	if !chk {
		t.Error("invalid flat config.")
	}
}

func TestGetAddrs(t *testing.T) {
	c := map[string]Layout{}
	eth0 := Ethernet{}
	ip, _ := ParseIP("192.168.0.1")
	eth0.Addresses = []IP{*ip}
	c["eth0"] = &eth0

	nc, err := GetConfig(c)
	if err != nil {
		t.Errorf("cannot get config. error: %s", err)
	}

	ncFlat := nc.Flatten()
	addrs := ncFlat["eth0"].GetAddrs()
	chkIP, _ := ParseIP("192.168.0.1")
	chk := false
	for _, addr := range addrs {
		if addr.String() == chkIP.String() {
			chk = true
		}
	}

	if !chk {
		t.Errorf("cannot get addrs. error: %s", err)
	}
}

func TestSetAddrs(t *testing.T) {
	c := map[string]Layout{}
	eth0 := Ethernet{}
	ip, _ := ParseIP("192.168.0.1")
	eth0.Addresses = []IP{*ip}
	c["eth0"] = &eth0

	nc, err := GetConfig(c)
	if err != nil {
		t.Errorf("cannot get config. error: %s", err)
	}

	ncFlat := nc.Flatten()
	newIP, _ := ParseIP("10.234.5.7")
	ncFlat["eth0"].UpdateAddrs([]IP{*newIP})
	chk := false
	addrs := ncFlat["eth0"].GetAddrs()
	for _, addr := range addrs {
		if addr.String() == newIP.String() {
			chk = true
		}
		// ips should be updated.
		if addr.String() == ip.String() {
			chk = false
		}
	}

	if !chk {
		t.Errorf("cannot set addrs. error: %s", err)
	}
}

func TestReadFile(t *testing.T) {
	file, err := os.CreateTemp("", "testreadfile")
	if err != nil {
		t.Errorf("cannot create temp file. error: %s", err)
	}
	err = os.WriteFile(file.Name(), []byte(config), os.FileMode(int(644)))
	if err != nil {
		t.Errorf("cannot create temp file. error: %s", err)
	}

	nc, err := ReadFile(file.Name())
	if err != nil {
		t.Errorf("cannot read config file. error: %s", err)
	}

	if v, ok := nc.Network.Vlans["vlan100"]; !ok {
		t.Error("invalid config")
	} else {
		chk := false
		chkip := "8.8.8.8"
		for _, a := range v.Nameservers.Addresses {
			if a.String() == chkip {
				chk = true
			}
		}
		if !chk {
			t.Error("invalid config")
		}
	}
}

func TestAddBond(t *testing.T) {
	nc, _ := GetConfig(nil)
	// invalid config interface
	err := nc.AddBond("test", "", []string{})
	if err != nil {
		if !strings.Contains(err.Error(), "invalid interface") {
			t.Errorf("invalid error: %s", err)
		}
	} else {
		t.Error("failed to properly capture error")
	}

	// invalid bonding name
	err = nc.AddBond("", "test", []string{})
	if err != nil {
		if !strings.Contains(err.Error(), "invalid interface") {
			t.Errorf("invalid error: %s", err)
		}
	} else {
		t.Error("failed to properly capture error")
	}

	// empty interfaces
	err = nc.AddBond("test", "test", []string{})
	if err != nil {
		if !strings.Contains(err.Error(), "empty interface") {
			t.Errorf("invalid error: %s", err)
		}
	} else {
		t.Error("failed to properly capture error")
	}

	err = nc.AddBond("test", "test", []string{"test2"})
	if err != nil {
		if !strings.Contains(err.Error(), "included within the interface") {
			t.Errorf("invalid error: %s", err)
		}
	} else {
		t.Error("failed to properly capture error")
	}

	c := map[string]Layout{}
	ip, _ := ParseIP("192.168.1.100")
	eth0 := &Ethernet{}
	eth0.Addresses = []IP{*ip}
	c["eth0"] = eth0

	nc, _ = GetConfig(c)
	// eth1 not found
	err = nc.AddBond("bond0", "eth0", []string{"eth0", "eth1"})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("cannot add bond: %s", err)
		}
	}

	c["eth1"] = &Ethernet{}
	nc, _ = GetConfig(c)
	err = nc.AddBond("bond0", "eth0", []string{"eth0", "eth1"})
	if err != nil {
		t.Errorf("cannot add bond: %s", err)
	}

	newBond, err := nc.GetBond("bond0")
	if err != nil {
		t.Errorf("cannot get bond: %s", err)
	}
	if !reflect.DeepEqual(newBond.Common, eth0.Common) {
		t.Errorf("invalid config, different from the configuration interface")
	}
}

func TestRemoveBond(t *testing.T) {
	c := map[string]Layout{}
	ip, _ := ParseIP("192.168.1.100")
	bond0 := &Bond{}
	bond0.Addresses = []IP{*ip}
	bond0.Interfaces = []string{"eth0", "eth1"}
	c["bond0"] = bond0
	nc, _ := GetConfig(c)
	nc.RemoveBond("bond0", "eth0")

	objs := nc.Flatten()

	if !reflect.DeepEqual(objs["eth0"], bond0.Dump()) {
		t.Errorf("invalid config, different from the bonding interface")
	}

	if _, ok := objs["bond0"]; ok {
		t.Error("remove bonding failed")
	}
}
