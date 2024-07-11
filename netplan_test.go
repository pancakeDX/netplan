package netplan

import (
	"net"
	"reflect"
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
								{IP: net.ParseIP("8.8.8.8")},
								{IP: net.ParseIP("8.8.4.4")},
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
	bytes, err := Write(&netplanConfig)
	if err != nil {
		t.Errorf("cannot write config. error: %v", err)
	}
	assert.Equal(t, string(bytes), config)
}

func TestGetConfig(t *testing.T) {
	c := map[string]interface{}{}
	eth0 := Ethernet{}
	eth1 := Ethernet{}
	ns := Nameservers{
		Addresses: []IP{
			{IP: net.ParseIP("8.8.8.8")},
			{IP: net.ParseIP("8.8.4.4")},
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
	c["eth0"] = eth0
	c["eth1"] = eth1
	c["bond0"] = bond0
	c["vlan100"] = vlan100

	nc, err := GetConfig(c)
	if err != nil {
		t.Errorf("cannot get config. error: %s", err)
	}

	eq := reflect.DeepEqual(*nc, netplanConfig)
	if !eq {
		t.Error("invalid config")
	}
}
