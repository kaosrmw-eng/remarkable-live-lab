//go:build tailscale

package main

import "net"

func (tm *TailscaleManager) PublicListener() (net.Listener, error) {
	return tm.server.ListenFunnel("tcp", ":8443")
}
