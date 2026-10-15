# Design: a 70s wooden television

The whole site plays on the screen of an old walnut TV set standing in a 70s
living room. The picture is warm, faded colour on a dark tube. Dark only.

## The set (`layout.templ`, "The television" in `assets/css/app.css`)

- `body.room`: striped brown wallpaper with an orange glow behind the set.
- `.tv__cabinet`: walnut wood grain (`--wood-light`, `--wood`, `--wood-dark`),
  rounded 28px, with tapered legs (`.tv__legs`) underneath.
- `.tv__bezel`: dark plastic surround (`--bezel`), curved 44px/52px. Inside it:
  - `.tv__screen`: the page itself. It scrolls on its own; the window doesn't.
  - `.tv__glass`: scanlines, dark curved edges, a reflection and a rolling band.
  - `.tv__static`: snow, shown on power-on and when changing channel.
  - `.tv__osd`: green channel number ("CH 02") top right, fades after 3s.
- `.tv__panel` (right side; under the screen on phones): "Wildvision" chrome
  plate, two chrome knobs, speaker grille, orange power light.

## Channels (`views/layout/channels.go`)

The nav pages are channels 1–6, in `channels`. The nav shows each number in
green. The **Channel** knob links to the next one, **Back** to the previous;
the channel knob is turned one notch (60°) per channel via `.tv--ch-N`.
Links with `data-channel` show a burst of static before going (`app.js`).
The set powers on (line → full picture) once per browser tab; after that
pages arrive with static clearing (`tv-warm`, set from `sessionStorage`).

## Colours (`:root`)

| Token | Hex | Use |
|---|---|---|
| `--bg` | `#1E1611` | the tube |
| `--surface` / `--surface-2` | `#2A2019` / `#33281F` | panels, cards |
| `--text` | `#F2E3C6` | cream text |
| `--muted` | `#C2AD8C` | secondary text |
| `--line` | `#4D3B2C` | borders |
| `--accent` | `#E8894A` | burnt orange: links, buttons, knob pointers |
| `--accent-2` | `#D9A93B` | mustard: eyebrow, ghost button, hover |
| `--teal` | `#6FB3A8` | faded teal |
| `--osd` | `#A8F07C` | on-screen channel numbers |

Animal groups: mammals orange, birds teal, reptiles mustard, sharks rust
`#C8553D`, other tan. IUCN badges keep the official Red List colours.

## Lettering

- Headings: Righteous (rounded, 70s), with a slight red/blue colour fringe (`--fringe`).
- Body: IBM Plex Mono.
- On-screen text (channel numbers, eyebrow, globe hint): VT323.

## Effects

- Photos are sepia-tinted and slightly washed out.
- Colour bars test strip above the home page title.
- Every animation is off under `prefers-reduced-motion`.

## Avoid

- Light mode. Neon green, the old cyberpunk look.
