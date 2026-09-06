//go:build tailscale

package main

import "net"

func (tm *TailscaleManager) PublicListener() (net.Listener, error) {
	return tm.server.ListenFunnel("tcp", ":8443")
}

// PrivateControlListener is reachable only inside the tailnet. HTTP is
// intentional: WireGuard supplies transport encryption, while using the stable
// 100.x address avoids mobile MagicDNS/public-Funnel resolution conflicts.
func (tm *TailscaleManager) PrivateControlListener() (net.Listener, error) {
	return tm.server.Listen("tcp", ":2003")
}
