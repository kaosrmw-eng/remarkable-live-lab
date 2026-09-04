# Security and operational limits

This is an experimental personal-device project, not a hardened multi-tenant service.

- Public viewing deliberately has no login. Starting sharing exposes the entire
  captured screen, including any sleep-screen calendar or dashboard content.
- Private controls require the tailnet route AND application credentials. Public
  and private handlers are separate; public requests cannot invoke control APIs.
- Leave `RK_TAILSCALE_FUNNEL=false`: that setting belongs to the private listener.
  The separate `RK_PUBLIC_VIEWER` feature exposes only the public listener on 8443.
- Do not use upstream `-unsafe` mode. Set unique credentials; never deploy example
  values. Keep the environment file, JWT signing material and tailnet state private.
- Frame discovery is a firmware-specific allocation heuristic. Exact size and
  mapping bounds reduce risk but are not a semantic guarantee of pixel identity.
  Zero or multiple candidates are rejected. Do not broaden firmware checks blindly.
- Capture failure must not stop network/control service; it should remain OFF
  until the owner successfully retries. No unattended arbitrary memory scan.
- Sleep guards stop current capture, not the server. The owner can explicitly
  start again while the device shows its sleep screen if networking is available.
- Service power-down stops this process and its embedded Tailscale connection.
  It does not turn off the tablet or other customizations. Waking alone does not
  restart the service; its installed boot unit is the recovery mechanism.
- Firmware updates can remove boot modifications or change capture compatibility.
- Eight concurrent connections were tested briefly, not eight remote users under
  sustained motion, poor Wi-Fi, or adversarial load. Public bandwidth is finite.
- Battery savings have not been quantified. Do not promise stock standby life.

Do not put passwords, screenshots of private notes, tokens or device identities
in public issues. For a suspected vulnerability, use a private channel to the
repository owner rather than publishing exploitable details with live endpoints.
