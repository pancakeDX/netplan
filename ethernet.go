package netplan

// Ethernet represents an ethernet interface
type Ethernet struct {
	Common `yaml:",inline"`
}

func (e *Ethernet) UpdateAddrs(ips []IP) {
	e.Addresses = ips
}

func (e *Ethernet) GetAddrs() []IP {
	return e.Addresses
}
