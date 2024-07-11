package netplan

import (
	"errors"
	"net"
)

type IP struct {
	IP   net.IP
	CIDR *net.IPNet
}

func (i *IP) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	// Attempt to resolve to IP
	if ip := net.ParseIP(str); ip != nil {
		i.IP = ip
		i.CIDR = nil
		return nil
	}

	// Attempt to resolve to CIDR
	ip, cidr, err := net.ParseCIDR(str)
	if err != nil {
		return errors.New("invalid IP or CIDR format")
	}

	i.IP = ip
	i.CIDR = cidr
	return nil
}

func (i IP) MarshalYAML() (interface{}, error) {
	if i.CIDR != nil {
		return i.CIDR.String(), nil
	}
	return i.IP.String(), nil
}

func (i IP) IsCIDR() bool {
	return i.CIDR != nil
}

func (i IP) String() string {
	if i.CIDR != nil {
		return i.CIDR.String()
	}
	return i.IP.String()
}

func ParseIP(str string) (*IP, error) {
	var i IP
	// Attempt to resolve to IP
	if ip := net.ParseIP(str); ip != nil {
		i.IP = ip
		i.CIDR = nil
		return &i, nil
	}

	// Attempt to resolve to CIDR
	ip, cidr, err := net.ParseCIDR(str)
	if err != nil {
		return nil, errors.New("invalid IP or CIDR format")
	}

	i.IP = ip
	i.CIDR = cidr
	return &i, nil
}
