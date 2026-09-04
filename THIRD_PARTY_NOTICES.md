# Attribution and provenance

The streaming implementation derives from Olivier Wulveryck's goMarkableStream:
https://github.com/owulveryck/goMarkableStream

Local upstream baseline: `8857534b585d1a4ee53dfbc32bd8ab1fac01082e`.
The original MIT license (copyright 2023 Olivier Wulveryck) is retained at the root
and under `stream/`. Changes for this experiment include the Move-specific bounded
allocation lookup, lifecycle/capture separation, sharing gates, timers, separate
public/private servers, owner UI, broadcast fanout and service power controls.

Go dependencies retain their own licenses; consult `stream/go.mod` and `go.sum`.
Notably, Tailscale/tsnet supplies private networking and public Funnel transport,
and klauspost/compress supplies compression used by the upstream delta protocol.
The browser's vendored fzstd decoder has its license under `stream/client/lib/`.

VDO.Ninja and Raspberry Ninja informed an earlier, separate prototype. That
prototype is described in the article brief but is not the current streaming
transport and its source is not included in this snapshot.

Rich Washburn's signature, coffee-cup artwork, personal content and screenshots
are not included or licensed for reuse by this code release. Product and company
names remain their respective owners' marks. No endorsement is implied.
