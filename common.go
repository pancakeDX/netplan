package netplan

// Common represents the common fields for various network configurations
type Common struct {
	Addresses   []IP        `yaml:"addresses,omitempty"`
	Gateway4    IP          `yaml:"gateway4,omitempty"`
	Nameservers Nameservers `yaml:"nameservers,omitempty"`
	Dhcp4       bool        `yaml:"dhcp4,omitempty"`
	Dhcp6       bool        `yaml:"dhcp6,omitempty"`
}

// Nameservers represents the nameservers configuration for a vlan interface
type Nameservers struct {
	Addresses []IP `yaml:"addresses,omitempty"`
}

type Layout interface {
	UpdateAddrs(ips []IP)
	GetAddrs() []IP

	UpdateNS(ips []IP)
	GetNS() []IP

	SetDhcp4(enable bool)

	SetGateway4(ip IP)
}
