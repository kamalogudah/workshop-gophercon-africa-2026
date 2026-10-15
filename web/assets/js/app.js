// Progressive enhancement for the site: every page works without this,
// it just doesn't move.
(() => {
	'use strict';

	document.documentElement.classList.add('js');

	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
	const finePointer = window.matchMedia('(pointer: fine)').matches;

	// Reads a colour token from the stylesheet, so the map and globe
	// follow the theme, dark mode included.
	const token = (name) => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

	const groupColour = (group) => ({
		'Mammals': token('--mammals'),
		'Birds': token('--birds'),
		'Reptiles': token('--reptiles'),
		'Sharks & rays': token('--sharks'),
	})[group] || token('--other');

	// Points are fetched once per URL, however many maps ask.
	const points = new Map();
	const loadPoints = (url) => {
		if (!points.has(url)) {
			points.set(url, fetch(url).then((r) => (r.ok ? r.json() : [])).catch(() => []));
		}
		return points.get(url);
	};

	/* ----------------------------------------------------------------------
	   The television: nav links and knobs change channel through a burst of
	   static, and the set only powers on once per visit.
	   ---------------------------------------------------------------------- */

	const television = () => {
		try {
			sessionStorage.setItem('tv-on', '1');
		} catch (e) {}

		const tv = document.querySelector('.tv');
		const screen = document.querySelector('[data-screen]');
		if (!tv) return;

		// Focus the screen so the arrow keys and Page Down scroll it.
		if (screen && !location.hash) screen.focus({ preventScroll: true });

		if (reducedMotion) return;

		document.addEventListener('click', (e) => {
			const link = e.target.closest('a[data-channel]');
			if (!link || e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
			e.preventDefault();
			tv.classList.add('is-switching');
			setTimeout(() => { location.href = link.href; }, 350);
		});

		// Coming back with the browser's back button shows the cached page, still snowy.
		window.addEventListener('pageshow', (e) => {
			if (e.persisted) tv.classList.remove('is-switching');
		});
	};

	/* ----------------------------------------------------------------------
	   3D tilt: cards lean towards the pointer, with a glare that follows it.
	   ---------------------------------------------------------------------- */

	const tilt = () => {
		if (reducedMotion || !finePointer) return;

		const max = 8; // degrees
		document.addEventListener('pointermove', (e) => {
			const el = e.target.closest?.('[data-tilt]');
			if (!el) return;
			const r = el.getBoundingClientRect();
			const x = (e.clientX - r.left) / r.width;
			const y = (e.clientY - r.top) / r.height;
			el.classList.add('is-tilting');
			el.style.setProperty('--ry', `${(x - 0.5) * 2 * max}deg`);
			el.style.setProperty('--rx', `${(0.5 - y) * 2 * max}deg`);
			el.style.setProperty('--gx', `${x * 100}%`);
			el.style.setProperty('--gy', `${y * 100}%`);
			el.style.setProperty('--glare', '1');
		});

		document.addEventListener('pointerout', (e) => {
			const el = e.target.closest?.('[data-tilt]');
			if (!el || el.contains(e.relatedTarget)) return;
			el.classList.remove('is-tilting');
			for (const p of ['--rx', '--ry', '--glare']) el.style.removeProperty(p);
		});
	};

	/* ----------------------------------------------------------------------
	   The weather card flips on click or Enter, for touch and keyboards.
	   ---------------------------------------------------------------------- */

	const flips = () => {
		for (const el of document.querySelectorAll('[data-flip]')) {
			const toggle = () => el.classList.toggle('is-flipped');
			el.addEventListener('click', toggle);
			el.addEventListener('keydown', (e) => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					toggle();
				}
			});
		}
	};

	/* ----------------------------------------------------------------------
	   Lightbox: photos open full size in a <dialog>.
	   ---------------------------------------------------------------------- */

	const lightbox = () => {
		const dialog = document.getElementById('lightbox');
		if (!dialog) return;
		const img = dialog.querySelector('.lightbox__img');
		const caption = dialog.querySelector('.lightbox__caption');

		document.addEventListener('click', (e) => {
			const trigger = e.target.closest('[data-lightbox]');
			if (!trigger) return;
			img.src = trigger.dataset.lightbox;
			img.alt = trigger.querySelector('img')?.alt || '';
			caption.textContent = trigger.dataset.caption || '';
			dialog.showModal();
		});

		// A click on the backdrop lands on the dialog itself.
		dialog.addEventListener('click', (e) => {
			if (e.target === dialog) dialog.close();
		});
	};

	/* ----------------------------------------------------------------------
	   Filters submit as soon as they change.
	   ---------------------------------------------------------------------- */

	const autosubmit = () => {
		for (const form of document.querySelectorAll('form[data-autosubmit]')) {
			form.addEventListener('change', () => form.requestSubmit());
		}
	};

	/* ----------------------------------------------------------------------
	   Maps, drawn with Leaflet on a canvas so thousands of dots stay quick.
	   ---------------------------------------------------------------------- */

	const tiles = () => window.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 18,
		attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
	});

	const escape = (s) => String(s).replace(/[&<>"']/g, (c) => ({
		'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
	})[c]);

	const maps = () => {
		if (!window.L) return;

		for (const el of document.querySelectorAll('[data-map]')) {
			const map = window.L.map(el, { preferCanvas: true, worldCopyJump: true }).setView([10, 20], 2);
			tiles().addTo(map);

			loadPoints(el.dataset.map).then((list) => {
				if (!list.length) return;
				const layer = window.L.featureGroup();
				for (const p of list) {
					const colour = groupColour(p.group);
					window.L.circleMarker([p.lat, p.lng], {
						radius: 6,
						color: '#fff',
						weight: 1,
						fillColor: colour,
						fillOpacity: 0.85,
					}).bindPopup(
						`<div class="map__popup"><strong>${escape(p.name)}</strong>` +
						`${escape(p.place)}<br>${escape(p.date)}<br>` +
						`<a href="/sightings/${escape(p.id)}">See the sighting →</a></div>`,
					).addTo(layer);
				}
				layer.addTo(map);
				map.fitBounds(layer.getBounds(), { padding: [24, 24], maxZoom: 8 });
			});
		}

		for (const el of document.querySelectorAll('[data-minimap]')) {
			const at = [Number(el.dataset.lat), Number(el.dataset.lng)];
			const map = window.L.map(el, { scrollWheelZoom: false }).setView(at, 9);
			tiles().addTo(map);
			window.L.marker(at, { title: el.dataset.label }).addTo(map);

			// The record's uncertainty, drawn as a circle around the point.
			const radius = Number(el.dataset.radius);
			if (radius > 0) {
				window.L.circle(at, { radius, color: token('--accent'), weight: 1, fillOpacity: 0.12 }).addTo(map);
			}
		}
	};

	/* ----------------------------------------------------------------------
	   Globe: every sighting on a spinning sphere, in plain canvas. Points
	   are projected orthographically, and those on the far side are
	   skipped; nearer ones are drawn larger and brighter for depth.

	   It spins up when the page loads, the dots drop in one by one, a
	   satellite circles it and random sightings ping. Clicking a dot turns
	   the globe to face it and projects a hologram of the sighting above.
	   ---------------------------------------------------------------------- */

	const globe = () => {
		const canvas = document.querySelector('[data-globe]');
		if (!canvas) return;
		const ctx = canvas.getContext('2d');
		const rad = Math.PI / 180;
		const tau = Math.PI * 2;
		const holo = canvas.parentElement.querySelector('[data-hologram]');

		let list = [];
		let lambda = -20 * rad; // longitude facing the viewer
		let phi = 12 * rad;     // tilt towards the viewer
		let spin = reducedMotion ? 0 : 0.0025;
		let dragging = null;
		let size = 0;
		let dpr = 1;
		let visible = [];       // what was drawn last frame, for hit testing
		let hovered = null;
		let selected = null;
		let turn = null;        // an eased rotation towards the selected point
		let loadedAt = 0;
		const started = performance.now();
		const pings = [];

		const resize = () => {
			dpr = Math.min(window.devicePixelRatio || 1, 2);
			size = canvas.clientWidth;
			canvas.width = size * dpr;
			canvas.height = size * dpr;
		};

		// project returns the point on screen and its depth, from 1 at
		// the front of the globe to 0 at its edge, or null if it's behind.
		const project = (lat, lng, r, c) => {
			const la = lat * rad;
			const lo = lng * rad + lambda;
			const x = Math.cos(la) * Math.sin(lo);
			const y = Math.sin(la);
			const z = Math.cos(la) * Math.cos(lo);
			// Tilt about the x axis.
			const y2 = y * Math.cos(phi) - z * Math.sin(phi);
			const z2 = y * Math.sin(phi) + z * Math.cos(phi);
			if (z2 < 0) return null;
			return { x: c + x * r, y: c - y2 * r, z: z2 };
		};

		const ease = (t) => 1 - Math.pow(1 - Math.min(Math.max(t, 0), 1), 3);

		// The satellite's orbit, tilted across the globe. Half of it runs
		// behind the sphere, so it's drawn in two passes.
		const orbit = (c, r, now, front) => {
			const rx = r * 1.14;
			const ry = r * 0.3;
			const tilt = -18 * rad;
			const from = front ? 0 : Math.PI;
			ctx.save();
			ctx.strokeStyle = token('--teal');
			ctx.globalAlpha = front ? 0.55 : 0.25;
			ctx.lineWidth = dpr;
			ctx.setLineDash([4 * dpr, 6 * dpr]);
			ctx.lineDashOffset = -now * 0.01 * dpr;
			ctx.beginPath();
			ctx.ellipse(c, c, rx, ry, tilt, from, from + Math.PI);
			ctx.stroke();
			ctx.setLineDash([]);

			const a = (now * 0.00035) % tau;
			if ((a < Math.PI) === front) {
				const ex = rx * Math.cos(a);
				const ey = ry * Math.sin(a);
				const x = c + ex * Math.cos(tilt) - ey * Math.sin(tilt);
				const y = c + ex * Math.sin(tilt) + ey * Math.cos(tilt);
				ctx.globalAlpha = front ? 1 : 0.4;
				ctx.fillStyle = token('--osd');
				ctx.shadowColor = token('--osd');
				ctx.shadowBlur = 10 * dpr;
				ctx.beginPath();
				ctx.arc(x, y, 2.5 * dpr, 0, tau);
				ctx.fill();
			}
			ctx.restore();
		};

		const ring = (x, y, radius, colour, alpha, width = 1.5) => {
			ctx.beginPath();
			ctx.arc(x, y, radius, 0, tau);
			ctx.strokeStyle = colour;
			ctx.globalAlpha = alpha;
			ctx.lineWidth = width * dpr;
			ctx.stroke();
			ctx.globalAlpha = 1;
		};

		const draw = (now) => {
			const c = (size * dpr) / 2;
			const r = c * 0.84;
			ctx.clearRect(0, 0, canvas.width, canvas.height);

			// The globe swells into place as the page loads.
			const grow = reducedMotion ? 1 : 0.85 + 0.15 * ease((now - started) / 900);
			ctx.save();
			ctx.translate(c, c);
			ctx.scale(grow, grow);
			ctx.translate(-c, -c);

			orbit(c, r, now, false);

			// The sphere, lit from the top left.
			const shade = ctx.createRadialGradient(c - r * 0.35, c - r * 0.4, r * 0.1, c, c, r);
			shade.addColorStop(0, token('--surface'));
			shade.addColorStop(1, token('--surface-2'));
			ctx.beginPath();
			ctx.arc(c, c, r, 0, tau);
			ctx.fillStyle = shade;
			ctx.fill();
			ctx.strokeStyle = token('--accent');
			ctx.lineWidth = dpr;
			ctx.shadowColor = token('--accent');
			// The rim breathes.
			ctx.shadowBlur = (16 + (reducedMotion ? 0 : Math.sin(now / 700) * 6)) * dpr;
			ctx.stroke();
			ctx.shadowBlur = 0;

			// Lines of latitude and longitude every 30 degrees.
			ctx.strokeStyle = token('--line');
			ctx.globalAlpha = 0.8;
			for (let lng = -180; lng < 180; lng += 30) line((t) => [t, lng], -90, 90);
			for (let lat = -60; lat <= 60; lat += 30) line((t) => [lat, t], -180, 180);
			ctx.globalAlpha = 1;

			function line(at, from, to) {
				ctx.beginPath();
				let open = false;
				for (let t = from; t <= to; t += 3) {
					const [lat, lng] = at(t);
					const p = project(lat, lng, r, c);
					if (!p) { open = false; continue; }
					if (open) ctx.lineTo(p.x, p.y); else ctx.moveTo(p.x, p.y);
					open = true;
				}
				ctx.stroke();
			}

			// The sightings, back to front so the nearest sit on top.
			visible = [];
			for (const s of list) {
				const p = project(s.lat, s.lng, r, c);
				if (!p) continue;
				// Each dot drops in after its own short delay.
				const k = reducedMotion ? 1 : ease((now - loadedAt - s.delay) / 500);
				if (k <= 0) continue;
				visible.push({ p, s, k });
			}
			visible.sort((a, b) => a.p.z - b.p.z);
			// Overlapping dots add up, so busy places glow.
			ctx.globalCompositeOperation = 'lighter';
			for (const { p, s, k } of visible) {
				const focus = s === hovered || s === selected;
				ctx.beginPath();
				ctx.arc(p.x, p.y, (1 + p.z * 2.2) * dpr * k * (focus ? 1.8 : 1), 0, tau);
				ctx.fillStyle = groupColour(s.group);
				ctx.globalAlpha = (0.25 + p.z * 0.75) * k;
				ctx.fill();
			}
			ctx.globalAlpha = 1;
			ctx.globalCompositeOperation = 'source-over';

			// Pings: rings that spread from a sighting and fade.
			for (let i = pings.length - 1; i >= 0; i--) {
				const t = (now - pings[i].at) / 1600;
				if (t >= 1) { pings.splice(i, 1); continue; }
				const p = project(pings[i].s.lat, pings[i].s.lng, r, c);
				if (p) ring(p.x, p.y, (2 + t * 16) * dpr, groupColour(pings[i].s.group), (1 - t) * p.z);
			}

			// The hovered dot gets its name.
			if (hovered && hovered !== selected) {
				const p = project(hovered.lat, hovered.lng, r, c);
				if (p) {
					ring(p.x, p.y, 7 * dpr, token('--text'), 0.8, 1);
					ctx.font = `${18 * dpr}px ${token('--font-osd')}`;
					ctx.fillStyle = token('--text');
					ctx.textAlign = p.x > c ? 'right' : 'left';
					ctx.fillText(hovered.name, p.x + (p.x > c ? -12 : 12) * dpr, p.y - 8 * dpr);
				}
			}

			orbit(c, r, now, true);
			ctx.restore();

			// The selected dot: a target lock and the projector's beam.
			if (selected) {
				ctx.save();
				ctx.translate(c, c);
				ctx.scale(grow, grow);
				ctx.translate(-c, -c);
				const p = project(selected.lat, selected.lng, r, c);
				if (p) {
					const holoColour = token('--holo');
					const t = (now % 1400) / 1400;
					ring(p.x, p.y, (4 + t * 14) * dpr, holoColour, 1 - t);
					ring(p.x, p.y, 6 * dpr, holoColour, 0.9);
					beam(p, holoColour, now);
				}
				ctx.restore();
			}
		};

		// beam is a cone of light from the selected dot up to the bottom
		// edge of the hologram's panel.
		const beam = (p, colour, now) => {
			if (holo.hidden) return;
			const panel = holo.querySelector('.hologram__panel').getBoundingClientRect();
			const box = canvas.getBoundingClientRect();
			const top = (panel.bottom - box.top) * dpr;
			if (top >= p.y - 8 * dpr) return;
			const left = (panel.left - box.left + 16) * dpr;
			const right = (panel.right - box.left - 16) * dpr;
			const g = ctx.createLinearGradient(0, p.y, 0, top);
			const flicker = reducedMotion ? 0.3 : 0.24 + Math.sin(now / 60) * 0.03 + Math.random() * 0.04;
			g.addColorStop(0, colour);
			g.addColorStop(1, 'transparent');
			ctx.globalCompositeOperation = 'lighter';
			ctx.globalAlpha = flicker;
			ctx.fillStyle = g;
			ctx.beginPath();
			ctx.moveTo(p.x - 2 * dpr, p.y);
			ctx.lineTo(left, top);
			ctx.lineTo(right, top);
			ctx.lineTo(p.x + 2 * dpr, p.y);
			ctx.closePath();
			ctx.fill();
			ctx.globalAlpha = 1;
			ctx.globalCompositeOperation = 'source-over';
		};

		// nearest is the drawn dot under the pointer, if there is one.
		const nearest = (e) => {
			const box = canvas.getBoundingClientRect();
			const x = (e.clientX - box.left) * dpr;
			const y = (e.clientY - box.top) * dpr;
			let best = null;
			let reach = (12 * dpr) ** 2;
			for (const { p, s } of visible) {
				const d = (p.x - x) ** 2 + (p.y - y) ** 2;
				if (d < reach) { reach = d; best = s; }
			}
			return best;
		};

		/* The hologram. */

		const fill = (key, value) => {
			const el = holo.querySelector(`[data-holo="${key}"]`);
			if (el) el.textContent = value || '';
		};

		const open = (s) => {
			selected = s;
			fill('group', s.group);
			fill('name', s.name);
			fill('scientific', s.scientific);
			fill('place', [s.flag, s.place].filter(Boolean).join(' '));
			fill('date', s.date);
			fill('weather', [s.weather, s.temperature].filter(Boolean).join(' · '));
			fill('status', s.status ? `${s.status} · ${s.statusLabel}` : '');
			holo.querySelector('[data-holo-row="weather"]').hidden = !s.weather;
			holo.querySelector('[data-holo-row="status"]').hidden = !s.status;
			holo.querySelector('[data-holo="link"]').href = `/sightings/${encodeURIComponent(s.id)}`;
			const photo = holo.querySelector('[data-holo="photo"]');
			photo.hidden = !s.photo;
			if (s.photo) photo.querySelector('img').src = s.photo;
			holo.style.setProperty('--tone', groupColour(s.group));

			// Replay the projection each time, even if it's already open.
			holo.hidden = false;
			holo.classList.remove('is-closing', 'is-open');
			void holo.offsetWidth;
			holo.classList.add('is-open');

			// Turn the globe so the dot faces the viewer, a little below
			// the middle, leaving room above for the hologram.
			let to = -s.lng * rad;
			to = lambda + ((((to - lambda) % tau) + tau + Math.PI) % tau) - Math.PI;
			const tilt = Math.max(-1.5, Math.min(1.5, s.lat * rad + 0.5));
			if (reducedMotion) {
				lambda = to;
				phi = tilt;
			} else {
				turn = { at: performance.now(), from: [lambda, phi], to: [to, tilt] };
			}
		};

		const close = () => {
			if (!selected) return;
			selected = null;
			turn = null;
			holo.classList.remove('is-open');
			holo.classList.add('is-closing');
			setTimeout(() => {
				if (!selected) holo.hidden = true;
			}, reducedMotion ? 0 : 250);
		};

		holo.querySelector('[data-hologram-close]').addEventListener('click', close);

		// The hologram tilts towards the pointer, and settles back when it
		// leaves.
		const panel = holo.querySelector('.hologram__panel');
		if (!reducedMotion) {
			panel.addEventListener('pointermove', (e) => {
				const box = panel.getBoundingClientRect();
				const x = (e.clientX - box.left) / box.width - 0.5;
				const y = (e.clientY - box.top) / box.height - 0.5;
				panel.classList.add('is-tilting');
				panel.style.setProperty('--holo-ry', `${x * 24}deg`);
				panel.style.setProperty('--holo-rx', `${-y * 18}deg`);
			});
			panel.addEventListener('pointerleave', () => {
				panel.classList.remove('is-tilting');
				panel.style.removeProperty('--holo-ry');
				panel.style.removeProperty('--holo-rx');
			});
		}
		document.addEventListener('keydown', (e) => {
			if (e.key === 'Escape') close();
		});

		// Only spin while the globe is on screen.
		let onScreen = true;
		new IntersectionObserver(([entry]) => { onScreen = entry.isIntersecting; }).observe(canvas);

		let lastPing = 0;
		const frame = (now) => {
			if (onScreen) {
				if (turn) {
					const k = ease((now - turn.at) / 900);
					lambda = turn.from[0] + (turn.to[0] - turn.from[0]) * k;
					phi = turn.from[1] + (turn.to[1] - turn.from[1]) * k;
					if (k >= 1) turn = null;
				} else if (!dragging && !selected) {
					// Spin up fast on load, then settle to a cruise.
					lambda += spin * (1 + 10 * Math.exp(-(now - started) / 700));
				}
				if (!reducedMotion && visible.length && now - lastPing > 700) {
					lastPing = now;
					pings.push({ s: visible[Math.floor(Math.random() * visible.length)].s, at: now });
				}
				draw(now);
			}
			requestAnimationFrame(frame);
		};

		canvas.addEventListener('pointerdown', (e) => {
			turn = null;
			dragging = { x: e.clientX, y: e.clientY, lambda, phi, moved: false };
			canvas.setPointerCapture(e.pointerId);
		});
		canvas.addEventListener('pointermove', (e) => {
			if (!dragging) {
				hovered = nearest(e);
				canvas.style.cursor = hovered ? 'pointer' : '';
				return;
			}
			const dx = e.clientX - dragging.x;
			const dy = e.clientY - dragging.y;
			if (Math.abs(dx) + Math.abs(dy) > 4) dragging.moved = true;
			const k = Math.PI / size;
			lambda = dragging.lambda + dx * k;
			phi = Math.max(-1.5, Math.min(1.5, dragging.phi + dy * k));
		});
		canvas.addEventListener('pointerup', (e) => {
			// A press that didn't drag is a click: open the dot under it,
			// or close the hologram if there isn't one.
			if (dragging && !dragging.moved) {
				const s = nearest(e);
				if (s) open(s); else close();
			}
			dragging = null;
		});
		canvas.addEventListener('pointercancel', () => { dragging = null; });
		canvas.addEventListener('pointerleave', () => { hovered = null; });

		new ResizeObserver(resize).observe(canvas);
		resize();
		loadPoints(canvas.dataset.globe).then((p) => {
			// Copies, so the delays don't leak into the map's points.
			list = p.map((s) => ({ ...s, delay: Math.random() * 1400 }));
			loadedAt = performance.now();
		});
		requestAnimationFrame(frame);
	};

	television();
	tilt();
	flips();
	lightbox();
	autosubmit();
	maps();
	globe();
})();
