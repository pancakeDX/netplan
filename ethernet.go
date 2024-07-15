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

func (e *Ethernet) UpdateNS(ips []IP) {
	e.Nameservers.Addresses = ips
}

func (e *Ethernet) GetNS() []IP {
	return e.Nameservers.Addresses
}
