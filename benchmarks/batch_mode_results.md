# Batch Mode Benchmark Results

**Date:** 2026-02-18
**Platform:** darwin/arm64 (Apple M3 Max)
**Go:** 1.24+, Ebitengine v2

All benchmarks run with `-benchmem -count=3`.

---

## Single Atlas Page (Ideal Case for Coalesced)

All 10,000 sprites share the same magenta placeholder page — one giant batch.

### Immediate

| Benchmark | Iters | ns/op | B/op | allocs/op |
|-----------|-------|-------|------|-----------|
| Static | 296 | 4,147,360 | 5,792,478 | 30,000 |
| Static | 301 | 3,928,046 | 5,765,517 | 30,000 |
| Static | 309 | 3,881,733 | 5,702,951 | 30,000 |
| Rotating | 308 | 3,982,118 | 5,790,107 | 30,000 |
| Rotating | 298 | 3,928,480 | 5,360,000 | 30,000 |
| Rotating | 290 | 3,936,593 | 5,360,000 | 30,000 |
| AlphaVarying | 301 | 3,985,801 | 6,219,643 | 30,000 |
| AlphaVarying | 315 | 3,973,152 | 6,386,808 | 30,000 |
| AlphaVarying | 319 | 3,938,623 | 5,360,000 | 30,000 |

### Coalesced

| Benchmark | Iters | ns/op | B/op | allocs/op |
|-----------|-------|-------|------|-----------|
| Static | 607 | 1,983,308 | 2,171,200 | 3 |
| Static | 579 | 1,970,207 | 2,171,200 | 3 |
| Static | 604 | 1,999,392 | 2,171,200 | 3 |
| Rotating | 576 | 2,091,525 | 2,171,200 | 3 |
| Rotating | 579 | 2,088,487 | 2,171,200 | 3 |
| Rotating | 573 | 2,080,648 | 2,171,200 | 3 |
| AlphaVarying | 591 | 2,037,734 | 2,171,200 | 3 |
| AlphaVarying | 592 | 2,036,248 | 2,171,200 | 3 |
| AlphaVarying | 591 | 2,026,544 | 2,171,200 | 3 |

### Summary (Single Page)

| Variant | Immediate avg ns/op | Coalesced avg ns/op | Speedup | Allocs |
|---------|--------------------|--------------------|---------|--------|
| Static | ~3,986K | ~1,984K | **2.0x faster** | 30,000 -> 3 |
| Rotating | ~3,949K | ~2,087K | **1.9x faster** | 30,000 -> 3 |
| AlphaVarying | ~3,966K | ~2,034K | **1.9x faster** | 30,000 -> 3 |

---

## Particle Draw (1,000 particles, single page)

| Mode | Iters | ns/op | B/op | allocs/op |
|------|-------|-------|------|-----------|
| Immediate | 5,774 | 230,966 | 606,022 | 3,000 |
| Immediate | 4,686 | 274,536 | 643,850 | 3,000 |
| Immediate | 3,910 | 270,347 | 536,000 | 3,000 |
| Coalesced | 23,746 | 48,998 | 221,504 | 3 |
| Coalesced | 24,262 | 48,782 | 221,504 | 3 |
| Coalesced | 23,529 | 50,530 | 221,504 | 3 |

### Summary (Particles)

| Mode | avg ns/op | allocs/op | Speedup |
|------|-----------|-----------|---------|
| Immediate | ~259K | 3,000 | - |
| Coalesced | ~49K | 3 | **5.3x faster** |

---

## Multi-Page (10K sprites, 4 atlas pages 128x128)

Sprites cycle through pages, producing batch runs of ~2,500.

| Mode | Iters | ns/op | B/op | allocs/op |
|------|-------|-------|------|-----------|
| Immediate | 308 | 3,935,484 | 5,775,627 | 30,000 |
| Immediate | 325 | 3,712,142 | 5,735,572 | 30,000 |
| Immediate | 326 | 3,684,651 | 5,685,067 | 30,000 |
| Coalesced | 342 | 3,517,001 | 5,844,190 | 30,000 |
| Coalesced | 351 | 3,534,352 | 5,949,731 | 30,000 |
| Coalesced | 349 | 3,543,340 | 6,101,412 | 30,000 |

### Summary (Multi-Page)

| Mode | avg ns/op | Speedup |
|------|-----------|---------|
| Immediate | ~3,777K | - |
| Coalesced | ~3,531K | **~7% faster** |

---

## Mixed (5K sprites + 5 emitters x 200 particles, 2 pages, mixed blend)

Emitters interleaved with sprites — sprite->particle->sprite transitions force flushes.

| Mode | Iters | ns/op | B/op | allocs/op |
|------|-------|-------|------|-----------|
| Immediate | 615 | 2,046,338 | 3,216,000 | 18,000 |
| Immediate | 588 | 2,097,359 | 3,766,076 | 18,000 |
| Immediate | 615 | 2,080,587 | 3,216,000 | 18,000 |
| Coalesced | 645 | 2,042,029 | 3,537,554 | 15,015 |
| Coalesced | 561 | 1,905,841 | 2,910,720 | 15,015 |
| Coalesced | 627 | 1,958,543 | 2,910,720 | 15,015 |

### Summary (Mixed)

| Mode | avg ns/op | avg allocs/op | Speedup |
|------|-----------|---------------|---------|
| Immediate | ~2,075K | 18,000 | - |
| Coalesced | ~1,969K | 15,015 | **~5% faster, 17% fewer allocs** |

---

## Worst Case (10K sprites alternating pages every command)

Batch runs of length 1 — coalesced has vertex-building overhead with zero batching benefit.

| Mode | Iters | ns/op | B/op | allocs/op |
|------|-------|-------|------|-----------|
| Immediate | 258 | 4,349,719 | 5,360,000 | 30,000 |
| Immediate | 278 | 4,280,473 | 5,360,001 | 30,000 |
| Immediate | 285 | 4,029,390 | 5,360,000 | 30,000 |
| Coalesced | 289 | 4,693,527 | 7,545,961 | 30,000 |
| Coalesced | 246 | 4,247,495 | 5,360,000 | 30,000 |
| Coalesced | 280 | 4,292,046 | 5,360,001 | 30,000 |

### Summary (Worst Case)

| Mode | avg ns/op | Speedup |
|------|-----------|---------|
| Immediate | ~4,220K | - |
| Coalesced | ~4,411K | **~5% slower** (expected) |

---

## Real-World Atlas (10K sprites, 2x 4096x4096 pages, runs of 1000)

Two full-size atlas pages like real TexturePacker output. Sprites grouped in runs of 1,000
on alternating pages (10 page swaps total) — simulates a tilemap with two spritesheets.

| Mode | Iters | ns/op | B/op | allocs/op |
|------|-------|-------|------|-----------|
| Immediate | 271 | 4,296,466 | 5,832,373 | 30,000 |
| Immediate | 292 | 4,016,107 | 5,778,016 | 30,000 |
| Immediate | 297 | 4,051,470 | 5,716,807 | 30,000 |
| Coalesced | 621 | 1,997,478 | 2,215,040 | 30 |
| Coalesced | 606 | 1,990,823 | 2,215,040 | 30 |
| Coalesced | 598 | 1,991,631 | 2,215,040 | 30 |

### Summary (Real-World Atlas)

| Mode | avg ns/op | avg allocs/op | Speedup |
|------|-----------|---------------|---------|
| Immediate | ~4,121K | 30,000 | - |
| Coalesced | ~1,993K | 30 | **2.1x faster, 1000x fewer allocs** |

---

## Key Takeaways

1. **Single page (ideal):** Coalesced is ~2x faster and eliminates virtually all per-frame allocations (30,000 -> 3).
2. **Particles:** Coalesced is ~5x faster for particle rendering.
3. **Multi-page realistic (128x128):** Coalesced still wins by ~7% even with 4 page breaks.
4. **Mixed scenes:** Coalesced ~5% faster with 17% fewer allocations despite sprite/particle interleaving.
5. **Real-world atlas (4096x4096):** Coalesced is **2.1x faster** with 1000x fewer allocs. Runs of 1,000 sprites per page swap are very batchable.
6. **Worst case (every sprite breaks batch):** Coalesced ~5% slower — acceptable regression for a pathological case that doesn't occur in practice.
