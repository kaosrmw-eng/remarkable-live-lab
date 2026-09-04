# reMarkable Live Lab

A reMarkable Paper Pro Move experiment by Rich Washburn: from a branded sleep
screen to a cloud-fed Daily Board and a tablet-hosted live presentation system.

**Experimental, firmware-specific source—not a universal installer or an official
reMarkable product.** Tested on Paper Pro Move (Chiappa), firmware **3.28.0.169**.
Back up your device and understand recovery before installing custom software.

## What is here

- `stream/`: adapted goMarkableStream source, private owner controls, anonymous
  public viewing, shared multi-viewer capture, session timers and service power-down.
- `device/`: the tested service definition/installer and the Daily Board image-pull
  package. Installers are tied to the development layout; read them before use.
- [Article background and detailed project story](docs/ARTICLE-BRIEF.md).
- [Setup and architecture](docs/SETUP.md).
- [Security and limitations](SECURITY.md).
- [Attribution and provenance](THIRD_PARTY_NOTICES.md).

This is a sanitized source snapshot. It intentionally excludes device credentials,
Tailscale state, calendar/mail data, screenshots of personal documents, brand
artwork, upstream sample private keys, generated binaries and the private Base44 app.
It is not a full backup or a turnkey clone of Rich's cloud dashboard.

## Current streaming capabilities

- Control page over private Tailscale HTTPS plus login.
- Separate public Funnel listener: read-only viewer, fullscreen and screenshots.
- Copy the direct public viewer URL from the owner panel.
- Sharing timer: 5–60 minutes, or until stopped/asleep.
- Shared capture/encoding at 5 fps; up to 16 connections. Eight simultaneous
  public-relay streams tested for 15 seconds, mostly static content.
- Owner-controlled presentation rotation; **automatic tablet orientation is not solved**.
- Capture starts OFF, including after reboot. Power/cover events and suspend
  detection stop existing sharing. Owner can explicitly start again.
- Separate service shutdown, with confirmation; a full tablet restart brings back
  the installed service, not active capture.
- Control/network startup independent of framebuffer discovery. Start sharing
  performs bounded discovery and can report failure without crashing networking.

## Build and targeted tests

Go 1.26 is required by the inherited module.

```sh
cd stream
go test -race . ./internal/remarkable ./internal/stream
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags tailscale -trimpath -o ../dist/gomarkablestream-share .
```

The upstream module path is retained to preserve internal imports. The upstream
README under `stream/` describes upstream behavior, not all changes in this fork.
Some upstream tools, fixtures and sample assets are omitted from this snapshot;
the targeted tests above are the validation baseline, not a claim that every
upstream benchmark/example remains runnable.

## No always-on companion computer

During normal operation, the tablet runs capture, controls and its embedded
Tailscale client. Funnel carries public viewer traffic; a browser renders it.
The separate Daily Board uses a cloud image endpoint. A Mac was used for building,
installation and diagnostics, not as the runtime screen-streaming relay.

## Credits

Built on [Olivier Wulveryck's goMarkableStream](https://github.com/owulveryck/goMarkableStream),
with Tailscale/tsnet and upstream compression/rendering components. This project
does not claim to have invented the original capture or streaming implementation.
See the retained MIT license and third-party notices.
