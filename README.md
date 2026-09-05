# WTRTO - War Thunder Real-Time Overlay

Open-source HUD overlay and telemetry logger for War Thunder, inspired by
[WTRTI](https://github.com/MeSoftHorny/WTRTI).

WTRTO reads live telemetry from War Thunder's built-in local API
(`http://localhost:8111`) and renders a customizable, click-through
overlay on top of the game - no game files are modified.

## Status

Early development, but usable - see [Releases](https://github.com/seneaLL/WTRTO/releases) for builds.

## Goals

- Cross-platform: Windows, Linux, macOS (macOS support undecided - may be dropped)
- Transparent, always-on-top, click-through overlay window
- Configurable indicators (speed, altitude, fuel, WEP timer, etc.)
- Fully open source (MIT)

## Planned

- Per-vehicle profiles
- Flight logging to CSV

## Feedback

Questions, bug reports, suggestions: [Discord](https://discord.gg/m8dSTaZcN4)

## Building

Requires Go 1.23+.

```bash
go build ./cmd/wtrto
```

## License

MIT - see [LICENSE](LICENSE).
