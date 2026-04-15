# Time-travel debugging — plan

Captured from design session 2026-04-15. Implementation to follow in
small, independently shippable commits.

## Scope

State-snapshot time travel. Record `State[T](w)` after each dispatched
event; let the developer scrub back through history and view past UI.
Read-only scrub only — no branching, no edit-and-continue, no effect
replay.

Out of scope for v1: event-log replay, full frame capture, persistent
data structures, side-effect suppression, state inspector, diff view,
history export.

## Snapshot mechanism

User state opts in via:

```go
type Snapshotter interface {
    Snapshot() any   // deep-ish copy; framework stores as-is
    Restore(any)     // overwrite receiver from a prior Snapshot()
}
```

No reflection. No gob fallback in v1 (decide later). If state does not
implement `Snapshotter`, history is skipped silently.

## Capture policy

- Snapshot AFTER the event handler runs, under `w.mu`, synchronously.
- Trigger on click / key / wheel / focus / timer / custom events.
- Skip plain `MouseMove`; drags coalesce to one snap per ~16ms and a
  final snap on release.
- Each entry: `{snap T, when time.Time, cause string}`. Cause is
  built from event type + focused widget ID.

## Retention

- Generic `[]T` ring buffer parameterized on the window's state type.
- Byte cap, not count cap. Default `HistoryBytes = 64 MiB`; configurable
  via `WindowCfg.HistoryBytes`.
- Approximate size via optional `Size() int` on `Snapshotter`; fallback
  estimate otherwise.
- Evict oldest when over cap. Silent (decide later whether to warn).

## Replay semantics — read-only

- Entering time-travel mode freezes the app: widget input ignored,
  timers/animations paused, no new snapshots.
- Only the debug bar accepts input: scrub, step ±1, resume.
- "Resume" moves cursor to latest snapshot, unfreezes. History never
  truncated.
- Document: rewinding state does not un-do past side effects.

## Re-render

- After `Restore`, set dirty flag; event loop reruns view fn exactly
  like production. No layout/render-cmd caching.
- `StateMap` namespaces opt in via `NamespaceOpts{Snapshotable: true}`.
  Scroll and focus opt in; caches (SVG, tessellation) do not.
- `w.Now()` virtual clock: wall clock when live, snapshot timestamp
  during scrub. Users who want accurate clock-driven views switch from
  `time.Now()` to `w.Now()`.

## Debug UI — dedicated window

Second `Window` spawned by `w.DebugWindow()` (or auto when
`DebugTimeTravel` set). v1 contents:

- Timeline slider, one tick per snapshot, colored by cause.
- Step buttons `⏮ ◀ ▶ ⏭`.
- Keyboard: `←/→` step, `Home/End` jump, `Space` freeze, `Esc` resume.
- Cause label: "click btnSave @ 14:22:03.412".
- Counter: `347 / 1000 (42 MiB)`.
- Explicit freeze toggle button.

## Concurrency

- Ring buffer owns a `sync.RWMutex`. Push = brief write; debug view =
  read. Independent of either window's `w.mu`.
- `Restore` is posted as an event to the app window. Debug window never
  calls `Restore` directly. No cross-window lock cycles.
- Freeze is `atomic.Bool`; handlers check at dispatch entry and drop
  events except `restoreEvent` and resume.
- Hot path when disabled: single `if w.history == nil` short-circuit.
  Zero extra allocations.

## Public API surface

```go
// WindowCfg
DebugTimeTravel bool   // opt-in; default false
HistoryBytes    int64  // byte cap; default 64 MiB when DebugTimeTravel

// Window
(w *Window) Now() time.Time
(w *Window) Checkpoint(cause string)
(w *Window) DebugWindow() *Window

// User-implemented
type Snapshotter interface {
    Snapshot() any
    Restore(any)
}

// Namespace registration (widget authors)
func RegisterNamespace(key nsKey, opts NamespaceOpts)
type NamespaceOpts struct{ Snapshotable bool }
```

## Tests (all under `-race`)

1. Ring push, evict-by-bytes, cursor, wrap-around.
2. `Snapshot`/`Restore` round-trip equality.
3. Whitelisted `StateMap` namespace survives scrub; unmarked does not.
4. Freeze drops events except restore/resume.
5. Byte-cap eviction keeps total under limit.
6. `w.Now()` returns snapshot timestamp during scrub, wall clock live.
7. `DebugTimeTravel: false` ⇒ zero extra allocs (`AllocsPerRun`).
8. Full scrub cycle on `todo` example: drive 10 events, scrub to 3,
   assert view output matches.
9. Cross-window event posting: restoreEvent triggers re-render.
10. Deadlock regression: rapid scrubs under load with `-race` + deadline.
11. `AllocsPerRun` dispatch with history disabled == baseline.
12. `AllocsPerRun` dispatch with history enabled, documented count.

Benchmarks: `BenchmarkSnapshot`, `BenchmarkRestore`,
`BenchmarkDispatchWithHistory`.

Demo vehicle: `examples/todo` converted to use time travel.

## Landing order

1. `w.Now()` virtual clock. Tiny, zero-risk, unlocks deterministic
   tests.
2. `Snapshotter` interface + ring buffer + byte-cap eviction
   (internal; no UI).
3. Snapshot-on-event wiring + freeze atomic.
4. `Restore` + cross-window event posting (synthetic debug window
   for tests).
5. Debug window v1 UI.
6. `WindowCfg.DebugTimeTravel` auto-spawns debug window.
7. Namespace whitelist for `StateMap`; scroll + focus opt in.
8. Convert `examples/todo` to demo.
9. Changelog + docs.

## Unresolved

- API names: `Snapshotter` vs `TimeTravel`; `Now()` vs `VirtualNow()`.
- Gob fallback: drop, or keep unadvertised?
- Debug window geometry: remembered per-app, or always fresh?
- Global `F5` freeze from app window, or require debug-window focus?
- Keep `Checkpoint(cause)`, or drop (auto-snapshot covers it)?
- Ring overflow: silent evict, or one-time stderr warning?
- Multi-window apps: per-window ring (rec) or combined history?
