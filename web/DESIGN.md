# Design: neon city at night

Cyberpunk, dark only. Black streets, matrix-green neon, alert-red warnings.

## Colours (`:root` in `assets/css/app.css`)

| Token | Hex | Use |
|---|---|---|
| `--bg` | `#000000` | page |
| `--surface` / `--surface-2` | `#0C130E` / `#121C15` | panels, cards |
| `--text` | `#E0E0E0` | body text |
| `--muted` | `#94A3B8` | labels, secondary text |
| `--line` | `#1F3326` | borders |
| `--accent` | `#00FF41` | matrix green: links, primary button, glow, active nav |
| `--accent-2` | `#FF3333` | alert red: eyebrow, ghost button, `//` section prefix, horizon |
| `--glow` / `--glow-red` | | text/box glow for each accent |

Animal groups: mammals green, birds cyan `#00E5FF`, reptiles yellow `#F5F500`,
sharks red, other grey. IUCN badge colours stay the official Red List ones.

## Lettering

- Headings: Orbitron 700/900, uppercase, letter-spaced (`--font-display`).
- Everything else: JetBrains Mono (`--font`). Both from Google Fonts in the layout.

## Shape

- Radii are effectively square (`--radius: 2px`).
- Buttons have two cut corners (clip-path); card and tile photos have the
  top-right corner cut (`--cut: 14px`).

## Effects

- Scanlines over the whole screen, with a faint band rolling down (`body::after`).
- Grid floor behind the hero, moving towards a red horizon (`.hero__grid`).
- Glitch title: `.glitch` with a `data-text` copy, red and cyan slices jump
  every ~5s, plus a skew jolt (`.glitch-jolt`).
- Panels glow green on hover; card photos get a green scan sweep.
- Globe: green glowing rim, dots blend additively so busy places glow.
- Map tiles inverted and tinted green.
- All animation is switched off under `prefers-reduced-motion`.

## Logo

A green hexagon with a paw inside (red pad), inline SVG in the header and as
the favicon. The name reads `WILD.SIGHTINGS` with a red dot.

## Avoid

- Light mode.
- Emoji as icons (photo placeholders say `NO SIGNAL`).
