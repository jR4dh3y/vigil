/* Live CCTV clocks — every [data-clock] element shows real local time. */
const clocks = document.querySelectorAll<HTMLElement>("[data-clock]");

function tick() {
	const stamp = new Date().toLocaleTimeString("en-GB", { hour12: false });
	for (const el of clocks) {
		el.textContent = stamp;
	}
}

if (clocks.length > 0) {
	tick();
	setInterval(tick, 1000);
}

/* Copy-to-clipboard with confirm/fail feedback on the button itself. */
const copyButtons = document.querySelectorAll<HTMLButtonElement>("[data-copy]");

async function copyText(text: string): Promise<boolean> {
	try {
		await navigator.clipboard.writeText(text);
		return true;
	} catch {
		// Non-secure context fallback (e.g. LAN http://box:4321)
		const area = document.createElement("textarea");
		area.value = text;
		area.style.position = "fixed";
		area.style.opacity = "0";
		document.body.appendChild(area);
		area.select();
		const ok = document.execCommand("copy");
		area.remove();
		return ok;
	}
}

for (const btn of copyButtons) {
	const label = btn.innerHTML;
	btn.addEventListener("click", async () => {
		const text = btn.dataset.copy;
		if (!text) return;
		const ok = await copyText(text);
		btn.textContent = ok ? "Copied" : "Failed";
		setTimeout(() => {
			btn.innerHTML = label;
		}, 1600);
	});
}

/* Hero scroll choreography: the paper text zone lifts off the background,
   then the owl leads the DVR preview in from the right. */
const hero = document.querySelector<HTMLElement>("[data-hero]");
const heroText = document.querySelector<HTMLElement>("[data-hero-text]");
const heroImg = document.querySelector<HTMLElement>("[data-hero-img]");
const owlAnchor = document.querySelector<HTMLElement>("[data-hero-owl-anchor]");
const heroCarrier = document.querySelector<HTMLElement>("[data-hero-carrier]");
const carrierOwl = document.querySelector<HTMLElement>("[data-hero-carrier-owl]");
const heroPaper = document.querySelector<HTMLElement>("[data-hero-paper]");
const heroDvr = document.querySelector<HTMLElement>("[data-hero-dvr]");

if (hero && heroText && heroImg && owlAnchor && heroCarrier && carrierOwl && heroPaper && heroDvr) {
	const textLayer = heroText;
	const imageLayer = heroImg;
	const carrier = heroCarrier;
	const anchor = owlAnchor;
	const carrierMark = carrierOwl;
	const paper = heroPaper;
	const dvr = heroDvr;
	let carrierStartX = 0;
	let carrierStartY = 0;

	function clamp(value: number): number {
		return Math.min(1, Math.max(0, value));
	}

	function measureCarrierStart() {
		const previousTextTransform = textLayer.style.transform;
		const previousCarrierTransform = carrier.style.transform;
		textLayer.style.transform = "none";
		carrier.style.transform = "none";

		const anchorBounds = anchor.getBoundingClientRect();
		const owlBounds = carrierMark.getBoundingClientRect();
		carrierStartX = anchorBounds.left - owlBounds.left;
		carrierStartY = anchorBounds.top - owlBounds.top;

		textLayer.style.transform = previousTextTransform;
		carrier.style.transform = previousCarrierTransform;
	}

	function updateHero(text: HTMLElement, img: HTMLElement) {
		const textH = text.offsetHeight;
		const p = clamp(window.scrollY / Math.max(1, textH));
		const reveal = clamp((window.scrollY - textH * 0.12) / Math.max(1, textH * 0.88));
		const easedReveal = 1 - (1 - reveal) ** 3;
		text.style.transform = `translate3d(0, ${-window.scrollY}px, 0)`;
		img.style.transform = `scale(${(1.12 - p * 0.12).toFixed(3)})`;
		carrier.style.transform = `translate3d(${(carrierStartX * (1 - easedReveal)).toFixed(1)}px, ${(carrierStartY * (1 - easedReveal)).toFixed(1)}px, 0) rotate(${((1 - easedReveal) * 5).toFixed(2)}deg)`;
		paper.style.opacity = easedReveal.toFixed(3);
		dvr.style.opacity = easedReveal.toFixed(3);
		dvr.style.transform = `scale(${(0.96 + easedReveal * 0.04).toFixed(3)})`;
	}

	measureCarrierStart();

	if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
		carrier.style.transform = "none";
		paper.style.opacity = "1";
		dvr.style.opacity = "1";
		dvr.style.transform = "none";
	} else {
		updateHero(textLayer, imageLayer);
		let ticking = false;
		window.addEventListener(
			"scroll",
			() => {
				if (ticking) return;
				ticking = true;
				requestAnimationFrame(() => {
					updateHero(textLayer, imageLayer);
					ticking = false;
				});
			},
			{ passive: true },
		);
		window.addEventListener("resize", () => {
			measureCarrierStart();
			updateHero(textLayer, imageLayer);
		});
	}
}
