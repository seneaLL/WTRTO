# wt-wiki-scraper

Scrapes aircraft limits (`Max Speed Limit (IAS)`, `Mach Number Limit`, `G limit`, `Flap Speed Limit (IAS)`, `Gear Speed Limit (IAS)`) from wiki.warthunder.com and writes them out in the same JSON format as `../data/aircraft_limits.json`.

Standalone Go module (its own `go.mod`) - building the main app is unaffected.

## Running

```bash
go run .
```

Defaults:

- aircraft list is fetched from `https://wiki.warthunder.com/aviation?v=l`;
- output is written to `../data/aircraft_limits.generated.json` (the main file is never overwritten automatically - review and merge by hand);
- 300ms minimum delay between requests, concurrency 3, to stay polite to the wiki.

Flags:

```bash
go run . -limit 20                          # process only the first 20 aircraft (for a quick check)
go run . -only su_34,mig_23mld              # process specific slugs/ids only
go run . -concurrency 1 -delay 500ms        # set concurrecny and delay
go run . -out ../data/aircraft_limits.json  # write straight to the live file (careful, overwrites it)
```

## How it works

- The `data-unit-id` / `/unit/<slug>` URL on the list page matches the in-game aircraft identifier (`su_34`, `a-20g`, etc.) - the same value telemetry reports as `type` and the same key used in `aircraft_limits.json`. No separate mapping is needed.
- Folder groups on the list page (`class="wt-tree_group"`, e.g. bundled premium variants) have no `/unit/...` link of their own and are excluded from the slug list.
- If a unit page has no `Limits:` section at all (gliders, helicopters, old biplanes, etc.), it's skipped and logged as `SKIPPED`.

## Things to double-check manually

- **Flap Speed Limit.** The wiki sometimes labels values as `L / T / C` (Landing / Takeoff / Combat), sometimes `T / C`, sometimes gives a single unlabeled number. Our schema only has `takeoff_landing` and `combat`, so:
  - if both L and T are present, the **lower** of the two is used (conservative, so the warning fires earlier);
  - if the labels can't be matched to the values, the first/last value is used and a warning is logged;
  - if there's a single unlabeled value, it's applied to both fields, also with a warning.
    These cases are worth checking against the actual page.
- **Max Speed / Mach for variable-sweep-wing aircraft.** If the wiki gives two numbers for one limit, they're treated as `{min, max}` (smaller/larger) without tying them to a specific sweep angle - the same approximation already used for `su_34`/`mig_23mld` in the main file. Worth verifying, especially if the label info (`.game-unit_chars-info`) gives specific angles rather than just two numbers.
- **G limit.** Expects exactly two numbers in the string (`≈ -3/6 G`) - the smaller goes to `g_limit_neg`, the larger to `g_limit_pos`.
- Anything that couldn't be parsed cleanly is flagged in the log - review those before merging `aircraft_limits.generated.json` into the live `data/aircraft_limits.json`.
