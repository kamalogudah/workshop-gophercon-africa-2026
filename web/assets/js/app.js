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
	   ---------------------------------------------------------------------- */

	const globe = () => {
		const canvas = document.querySelector('[data-globe]');
		if (!canvas) return;
		const ctx = canvas.getContext('2d');
		const rad = Math.PI / 180;

		let list = [];
		let lambda = -20 * rad; // longitude facing the viewer
		let phi = 12 * rad;     // tilt towards the viewer
		let spin = reducedMotion ? 0 : 0.0025;
		let dragging = null;
		let size = 0;
		let dpr = 1;

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

		const draw = () => {
			const c = (size * dpr) / 2;
			const r = c * 0.9;
			ctx.clearRect(0, 0, canvas.width, canvas.height);

			// The sphere, lit from the top left.
			const shade = ctx.createRadialGradient(c - r * 0.35, c - r * 0.4, r * 0.1, c, c, r);
			shade.addColorStop(0, token('--surface'));
			shade.addColorStop(1, token('--surface-2'));
			ctx.beginPath();
			ctx.arc(c, c, r, 0, Math.PI * 2);
			ctx.fillStyle = shade;
			ctx.fill();
			ctx.strokeStyle = token('--accent');
			ctx.lineWidth = dpr;
			ctx.shadowColor = token('--accent');
			ctx.shadowBlur = 16 * dpr;
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
			const visible = [];
			for (const s of list) {
				const p = project(s.lat, s.lng, r, c);
				if (p) visible.push([p, s.group]);
			}
			visible.sort((a, b) => a[0].z - b[0].z);
			// Overlapping dots add up, so busy places glow.
			ctx.globalCompositeOperation = 'lighter';
			for (const [p, group] of visible) {
				ctx.beginPath();
				ctx.arc(p.x, p.y, (1 + p.z * 2.2) * dpr, 0, Math.PI * 2);
				ctx.fillStyle = groupColour(group);
				ctx.globalAlpha = 0.25 + p.z * 0.75;
				ctx.fill();
			}
			ctx.globalAlpha = 1;
			ctx.globalCompositeOperation = 'source-over';
		};

		// Only spin while the globe is on screen.
		let onScreen = true;
		new IntersectionObserver(([entry]) => { onScreen = entry.isIntersecting; }).observe(canvas);

		const frame = () => {
			if (onScreen) {
				if (!dragging) lambda += spin;
				draw();
			}
			requestAnimationFrame(frame);
		};

		canvas.addEventListener('pointerdown', (e) => {
			dragging = { x: e.clientX, y: e.clientY, lambda, phi };
			canvas.setPointerCapture(e.pointerId);
		});
		canvas.addEventListener('pointermove', (e) => {
			if (!dragging) return;
			const k = Math.PI / size;
			lambda = dragging.lambda + (e.clientX - dragging.x) * k;
			phi = Math.max(-1.2, Math.min(1.2, dragging.phi + (e.clientY - dragging.y) * k));
		});
		const release = () => { dragging = null; };
		canvas.addEventListener('pointerup', release);
		canvas.addEventListener('pointercancel', release);

		new ResizeObserver(resize).observe(canvas);
		resize();
		loadPoints(canvas.dataset.globe).then((p) => { list = p; });
		requestAnimationFrame(frame);
	};

	tilt();
	flips();
	lightbox();
	autosubmit();
	maps();
	globe();
})();
