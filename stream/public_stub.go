//go:build !tailscale

package main

import (
	"fmt"
	"net"
)

func (tm *TailscaleManager) PublicListener() (net.Listener, error) {
	return nil, fmt.Errorf("Tailscale unavailable")
}
