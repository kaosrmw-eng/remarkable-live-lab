# Architecture and setup notes

## Two independent paths

Daily Board: cloud data/rendering → authenticated PNG pull over Wi-Fi → image in
writable `/home` → sleep-screen display on the next transition.

Live sharing: tablet display process → bounded read-only capture → shared
delta/compression encoder → Tailscale Funnel → browser pixel renderer.

Owner: tailnet-connected browser → private control listener → login → start,
stop, rotation, timer, copy link, or service shutdown.

Private endpoint: `https://<device>.<tailnet>.ts.net/control` (443).
Phone-safe fallback: `http://<device-tailscale-100.x-ip>:2003/control`.
Public endpoint: `https://<device>.<tailnet>.ts.net:8443/`.
The same hostname resolves differently on private tailnet DNS versus public DNS.
Phone browser DNS/proxy settings can matter; a green device indicator does not
prove that a browser is using the private route. The numeric port-2003 fallback
avoids that ambiguity and is bound through tsnet, not Wi-Fi or Funnel. HTTP is
intentional for that tailnet-only route because WireGuard encrypts transport;
application login is still required. Do not bypass TLS warnings on the hostname.

## Requirements

- Paper Pro Move/Chiappa on the explicitly checked firmware 3.28.0.169.
- Developer Mode, backups, recovery knowledge and USB SSH access for installation.
- Go 1.26 on a development machine for ARM64 cross-compilation.
- Tailscale account, enrolled persistent device identity, HTTPS/MagicDNS and the
  appropriate Funnel permission. The private owner client also needs Tailscale.
- No Raspberry Pi, Windows box or always-on Mac required afterward.
- Daily Board separately needs an authenticated image endpoint; the private
  Base44 application and Google OAuth setup are not exported in this repository.

## Installation is manual and device-specific

1. Read all included scripts and verify firmware/model; do not use them unchanged
   on other devices. Preserve stock images and all existing customizations.
2. Build the ARM64 binary using the root README instructions.
3. Create `/home/root/gomarkable-move-test` on the tablet; copy the binary there
   as `gomarkablestream-share`. This historical staging name is retained by the unit.
4. Copy `device/gomarkable-private.env.example` as `gomarkable-private.env`, set
   unique credentials/hostname, and chmod 0600. Never put this filled file in Git.
5. Test the binary interactively and enroll its device using the Tailscale login
   flow. Confirm private login and public read-only boundaries before sharing.
   After login, the Security panel can replace the bootstrap password. The new
   credential is stored as a mode-0600 bcrypt hash and invalidates all sessions.
6. Only then consider the supplied installer. It briefly remounts `/` writable
   to add one systemd unit and a boot link under `/lib`, and restores read-only.
   `/etc` is volatile on the tested firmware, so ordinary enabling there is not
   a reliable persistent installation. Installer intentionally refuses to replace
   an existing unit; upgrades need a reviewed manual procedure.
7. Test a full reboot and first unlock. The device may not become reachable until
   its encrypted home and networking are available. Capture must start OFF.

The release copy removes upstream embedded sample TLS credentials; generate or
provide unique certificates for any non-tsnet TLS listener. Tailscale TLS uses
the enrolled device's certificates. The copy-link UI derives the viewer hostname
from the current private control hostname instead of embedding Rich's device URL.

## Daily Board package

`device/rich-daily-board/` contains updater, unit/timer and install/uninstall scripts.
The updater uses a narrow bearer token and conditional requests. Expected responses:
200 image/png with a 954×1696 image; 304 unchanged; 401 invalid token; 503 unavailable.
It validates the download before atomic replacement and avoids unchanged writes.
The existing sleep-image symlink/customization must be prepared separately; this
source snapshot does not include brand graphics or a complete sleep-screen installer.

## Recovery

To stop streaming infrastructure immediately, use authenticated Power down or
`systemctl stop gomarkable-share.service` over maintenance SSH. The app exits
successfully after owner power-down, so Restart=on-failure does not relaunch it.
To restore without reboot during maintenance: `systemctl start gomarkable-share.service`.
To remove startup, remove only the project's unit/link under a controlled writable
remount; restore read-only afterward. Preserve identity/secrets for rollback unless
you intentionally retire the installation. Full device recovery is a separate vendor
procedure, not replaced by these scripts.
