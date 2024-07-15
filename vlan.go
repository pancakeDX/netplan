package netplan

// Vlan represents a vlan interface
type Vlan struct {
	Common `yaml:",inline"`
	ID     int    `yaml:"id,omitempty"`
	Link   string `yaml:"link,omitempty"`
}

func (v *Vlan) UpdateAddrs(ips []IP) {
	v.Addresses = ips
}
func (v *Vlan) GetAddrs() []IP {
	return v.Addresses
}
