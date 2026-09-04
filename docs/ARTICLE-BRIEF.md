# From a branded sleep screen to a tablet-hosted presentation system

## Working headline ideas

- My reMarkable Became a Live Whiteboard—Without a Companion Computer
- The Little Linux Computer Inside My reMarkable
- From Coffee-Cup Branding to a Cloud-Connected Paper Dashboard

## The short version

Rich Washburn enabled Developer Mode on a reMarkable Paper Pro Move to explore
what the device could do beyond its normal writing interface. The initial goal
was deliberately small: replace the sleep screen with a clean, personal design.
That grew into a calendar-aware Daily Board connected to a cloud dashboard,
followed by a live screen-sharing system served from the tablet itself.

The current streaming system has two audiences: Rich uses a private control page
to start and stop a session, while viewers use an anonymous public link. The tablet
captures and distributes the display; Tailscale supplies connectivity and Funnel
public access. A Mac helped build, install and debug the project, but is not a
required runtime relay. No Raspberry Pi or Windows companion box is involved.

This was an iterative engineering experiment, not a finished consumer app. Its
most valuable lessons came from constraints: sleep-screen overlays, missing
multimedia runtimes, private-versus-public networking, real power tradeoffs, and
a reboot that changed the display process's memory layout.

## 1. Starting small: a personal sleep screen

The first concept treated the sleeping tablet as a physical identity object.
The visual direction became deliberately minimal: a white center, dark/red bands,
Rich's signature and solid red coffee cup, concise identifying text, a website
address and a large QR code. Busy website-style backgrounds and a lost-device
message were rejected. The brand mattered more than matching the whole website.

One small but important correction was using the supplied opaque coffee-cup
graphic instead of a website asset with a transparent cutout. Another was learning
that replacing one sleep image was not the entire display pipeline: a stock
carousel illustration could still appear over the custom screen. The installation
work accounted for those overlays and retained stock-image backups for reversal.

The panel artwork was prepared at 954×1696. Later, screen capture required a
different distinction: the internal pixel buffer has a padded 960-pixel stride.
That is a useful article detail—physical panel geometry and memory layout are
not necessarily identical.

## 2. Turning the sleep screen into a Daily Board

Once the visual identity worked, the information became more useful than a static
business card. The project evolved toward a compact daily view: upcoming events,
today's priorities, a brief tomorrow section and a small freshness indication.
The logo stayed at the top. Footer space became a stronger QR/dashboard entry
point rather than unnecessary agent status text.

Rich built the cloud dashboard in Base44 and made it private. The intended QR
destination shifted from the general website to the mobile dashboard: glance at
paper for a compact view, scan for the deeper interactive interface.

The device-side updater was designed to pull a rendered PNG from the cloud with
a narrow device token. It uses conditional requests (ETags), validates the image,
and installs changed content atomically. An unchanged response avoids rewriting
the file. Routine updates use writable home storage; they do not require remounting
the system partition each time. Updates are wake/network-dependent, not a promise
that the cloud can push into a radio that is fully asleep.

During development, Rich reported that calendar/board updates reached the device
and that the result was useful. The local repository contains the updater and its
contract; it does not contain a verified export of the entire current Base44 app.
For an article, describe that integration as part of this personal deployment,
not as a ready-to-install cloud backend supplied by this repository.

## 3. Where the persistent agent fits

Rich's persistent Base44 super-agent, Cora, was part of the broader plan. A pasted
agent report described a calendar-refresh function and two active workflows:
a morning refresh and a calendar-change trigger. That report also noted a coverage
limitation: the trigger watched the primary calendar, so other calendars could
depend on the scheduled refresh. Treat those as reported cloud configuration,
not independent guarantees about current provider behavior.

The broader idea is a paper-to-agent loop: selected notes leave the tablet, Cora
can process them in the cloud, and useful results return to the dashboard/board.
We discussed exporting documents into an agent intake location and eventually
automating capture. We have **not demonstrated a complete automatic pipeline that
scoops up every handwritten notebook and sends it to Cora**. That belongs in the
future-work section, not the finished-feature list.

Likewise, access to personal and business Google accounts at different layers
does not automatically merge all permissions. The application and agent need
explicit interfaces and scoped authorization between them.

## 4. Exploring VDO.Ninja before choosing the simpler route

Rich already hosted VDO.Ninja and knew its website embedding worked well. An
attractive goal was a tablet-native publisher that connected directly to that
existing deployment, without a desktop screen-capture bridge.

The tablet's runtime changed the engineering decision. Python, GStreamer and
FFmpeg were absent. Instead of treating a general Linux multimedia installer as
safe for the tablet, an isolated Go/Pion experiment was built. A synthetic,
pre-encoded VP8 pattern at 480×848 and 5 fps was transmitted from the tablet and
seen in the hosted browser viewer. Signaling compatibility required matching the
instance's stream-identifier/password-salt behavior.

That proved a transport experiment, **not complete native live-screen publishing
to VDO.Ninja**. Actual capture, encoding, reconnection, relayed connectivity and
power cost would still need engineering. The project therefore moved to
goMarkableStream for the usable live-screen feature.

## 5. The chosen streaming architecture

Olivier Wulveryck's goMarkableStream provided the starting point. The adapted
implementation reads the display process's pixel buffer, uses its delta/compression
protocol, and sends data to a browser renderer. It is not a conventional H.264 or
VP8 video broadcast and is not the official reMarkable Screen Share service.

The tablet runs a Go binary containing both the web server and Tailscale's tsnet
client. The owner control page is private on port 443; an independent read-only
viewer is exposed through Funnel on 8443. Keeping the handlers separate matters:
hiding buttons would not be enough if anonymous viewers could still call the APIs.

The public route has no start, stop, login or management endpoint. Tests checked
that attempts to reach those paths were rejected. Private mutations require an
authenticated session, and browser commands receive origin checks. Viewers can
see the screen, enter fullscreen and take a screenshot using the red camera icon.

Website embedding remains a next integration step. The current direct viewer is
the working destination, and the control panel can copy its URL. Do not describe
Wix/Bluehost embedding as already deployed by this work.

## 6. Making it feel like an actual presentation tool

Controls evolved from a combined test page into an owner-only surface:

- Start public sharing and Stop sharing.
- A 5–60-minute session duration in five-minute steps, default 30 minutes.
- A Permanent option meaning no session timer—not automatic capture after reboot.
- Presentation rotation controlled by the owner and reflected in the viewer.
- Copy public URL.
- Power down sharing service, with a separate confirmation.

Automatic orientation following is still unresolved. Physical accelerometer data
did not reliably identify the orientation used by the tablet's interface, and a
restricted D-Bus route was not bypassed. Manual owner rotation is the demonstrated
feature. This is an important correction to any story implying automatic rotation
is finished.

The initial upstream stream path admitted one viewer at a time. That was inadequate
for demonstrations to six or eight people. The public stream was replaced with
one shared capture/encoder feeding bounded per-viewer queues. New viewers receive
a full frame; slow connections disconnect instead of silently losing part of a
delta chain. A 16-connection ceiling gives headroom above the expected audience.

A live test through the public relay held eight connections for 15 seconds. Each
received 73 protocol frames without disconnecting. A browser also rendered the
real tablet display. The content was mostly static: this validates simultaneous
delivery, not a guarantee of eight flawless remote viewers under all Wi-Fi or
high-motion conditions. Empty deltas are protocol frames, not 73 visible changes.

## 7. Power: stop the picture or stop the service?

Rich noticed that the control page could remain reachable while the tablet showed
its sleep screen. An explicit Start could then share that displayed information.
For this presentation-oriented device, that behavior was useful rather than a
privacy emergency—but it raised the right question about battery consumption.

We separated three states: active sharing, capture stopped while controls remain
online, and the entire sharing service powered down. The last option exits the
process and its embedded Tailscale connection; it does not power down the tablet
or stop the Daily Board. With Restart=on-failure, a successful owner-requested exit
stays off. A full tablet restart launches the installed service again, capture OFF.
Simply pressing sleep/wake is not a service restart.

To reduce unnecessary work, idle page polling was slowed to 15 seconds. Hidden
tabs stop polling, and hidden viewers close their streaming connections. An idle
five-times-per-second suspend polling loop was removed; capture/output and
control operations retain suspend checks, while hardware input reads block.

The shutdown test verified no remaining service PID, a successful exit, Tailscale
shutdown, and no automatic restart. **Battery savings have not been measured.**
The defensible claim is reduced activity and an explicit off switch, not a battery
percentage or stock standby-life promise. Other tablet services can still run.

## 8. The reboot bug—and the strongest engineering lesson

The first full morning reboot exposed a weakness. Tailscale appeared offline,
but the real error was `allocation outside mapping` in framebuffer discovery.
The service had retried 38 times. It was exiting before networking could start.

The original Move adaptation assumed the frame allocation followed the last DRM
mapping. After reboot, that neighboring region was too small; the real frame-sized
allocation was elsewhere. The bounds check correctly rejected the read. Removing
the check would have been the wrong fix.

The replacement follows valid allocation headers within bounded anonymous mappings,
requires the expected page-rounded frame size, and rejects ambiguous candidates.
It also skips the allocation header when returning the pixel address. Model and
firmware checks remain because this is still a heuristic tied to implementation
details, not a stable vendor capture API.

More importantly, networking no longer depends on capture discovery. The control
panel starts independently. Pressing Start opens and locates the current display
buffer; a failure leaves controls online and allows an explicit retry.

After installation, a full reboot was performed. The service started automatically,
Tailscale reconnected, the control page came back with sharing OFF, and starting
sharing produced the correct live screen. Zero crash restarts were confirmed.
This is a concrete success to include in the article—and a useful explanation
of why failure isolation matters even in a small personal project.

## 9. What is finished, and what is not

Demonstrated: custom sleep-screen presentation; user-confirmed cloud Daily Board
updates; private/public streaming separation; real screen rendering; timer and
manual rotation; copy link; eight concurrent transport connections; explicit
service power-down; and successful reboot recovery after the capture fix.

Not established: quantified battery life, general support for other models or
firmware, automatic orientation, complete handwritten-note ingestion to Cora,
a deployed website embed, or production-grade sustained multi-user performance.
The phone's private-control TLS issue was consistent with taking a public DNS/relay
route instead of the private tailnet route; a confirmed phone-side resolution was
not recorded. Do not claim universal mobile connectivity based only on desktop tests.

## Suggested article structure and visuals

Open with the small request: “I just wanted my coffee cup and signature on the
sleep screen.” Show the finished board, then reveal the Linux computer underneath.
Walk through the no-companion-computer constraint, the VDO.Ninja experiment, the
chosen stream design and the private control/public audience distinction. Use
the reboot failure as the technical turning point. Finish with the agent loop
as the next experiment rather than an already completed capability.

Useful original visuals: the clean branded board using sample events; tablet and
browser side by side on a deliberately nonprivate sketch; the control panel; a
small architecture diagram; and the concise reboot error alongside the fix.
Replace real appointments, mail, notes, private URLs and tokens before publishing.

Suggested closing thought: the value was not turning an e-paper tablet into a
general-purpose laptop. It was keeping the writing experience intact while giving
the paper surface useful connections—to a cloud dashboard, an audience, and
eventually an agent.

## Sources and evidence

This brief summarizes local implementation, device logs/browser tests, and Rich's
reported deployment experience from the September 2–4, 2026 project work. It is
not a new audit of the cloud application. Primary technical starting points:

- https://github.com/owulveryck/goMarkableStream
- https://github.com/steveseguin/vdo.ninja
- https://github.com/steveseguin/raspberry_ninja
- https://tailscale.com/docs/features/tsnet
- https://tailscale.com/docs/features/tailscale-funnel

Attribute the upstream work clearly. The interesting contribution here is the
integration and iteration around this specific device and workflow—not a claim
that the underlying open-source projects were created from scratch.
