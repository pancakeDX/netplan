package netplan

import (
	"testing"

	"gotest.tools/assert"
)

func TestParseIP(t *testing.T) {
	ipStr := "192.168.0.1"
	ip, err := ParseIP(ipStr)
	if err != nil {
		t.Errorf("cannot parse ip. error: %s", err)
	}

	if ip.IsCIDR() {
		t.Error("Expected not to be a CIDR, but determined to be a CIDR.")
	}

	assert.Equal(t, ipStr, ip.String())
}

func TestParseCIDR(t *testing.T) {
	cidrStr := "10.123.7.0/24"
	cidr, err := ParseIP(cidrStr)
	if err != nil {
		t.Errorf("cannot parse cidr. error: %s", err)
	}

	if !cidr.IsCIDR() {
		t.Error("Expected to be a CIDR, but determined not to be a CIDR.")
	}

	assert.Equal(t, cidrStr, cidr.String())
}
