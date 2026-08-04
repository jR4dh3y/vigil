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
const carrierOriginal = carrierOwl?.querySelector<HTMLElement>('[data-hero-owl-state="original"]');
const carrierPointing = carrierOwl?.querySelector<HTMLElement>('[data-hero-owl-state="pointing"]');
const heroPaper = document.querySelector<HTMLElement>("[data-hero-paper]");
const heroDvr = document.querySelector<HTMLElement>("[data-hero-dvr]");
const heroPoints = document.querySelector<HTMLElement>("[data-hero-points]");
const heroPointItems = Array.from(
	heroPoints?.querySelectorAll<HTMLElement>("[data-hero-point]") ?? [],
);
const hangingOwl = document.querySelector<HTMLElement>("[data-hanging-owl]");
const installSection = document.querySelector<HTMLElement>("#install");

function clamp(value: number): number {
	return Math.min(1, Math.max(0, value));
}

function updateHangingOwlVisibility() {
	if (!hangingOwl || !installSection) return;
	const installIsReached = installSection.getBoundingClientRect().top <= window.innerHeight;
	hangingOwl.dataset.visible = String(installIsReached);
}

if (hangingOwl && installSection) {
	updateHangingOwlVisibility();
	window.addEventListener("scroll", updateHangingOwlVisibility, { passive: true });
	window.addEventListener("resize", updateHangingOwlVisibility);
}

function updateHeroPoints(progress: number) {
	// Leave a small lift at the end so long tags stay in the picture band above
	// the install/footer panel when it settles over the lower 75% of the viewport.
	const finalLift = 14;
	for (const point of heroPointItems) {
		const stagger = Number.parseFloat(point.dataset.dropStagger ?? "0");
		const itemProgress = clamp((progress - stagger) / Math.max(0.01, 1 - stagger));
		const offset = -100 + itemProgress * (100 - finalLift);
		point.style.setProperty("--drop-offset", `${offset.toFixed(1)}%`);
	}
}

if (
	hero &&
	heroText &&
	heroImg &&
	owlAnchor &&
	heroCarrier &&
	carrierOwl &&
	carrierOriginal &&
	carrierPointing &&
	heroPaper &&
	heroDvr
) {
	const textLayer = heroText;
	const imageLayer = heroImg;
	const carrier = heroCarrier;
	const anchor = owlAnchor;
	const carrierMark = carrierOwl;
	const originalOwl = carrierOriginal;
	const pointingOwl = carrierPointing;
	const paper = heroPaper;
	const dvr = heroDvr;
	let carrierStartX = 0;
	let carrierStartY = 0;

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
		// Start the carrier on the very first scroll frame so the text cannot
		// slide underneath a stationary owl before the app reveal begins.
		const reveal = clamp(window.scrollY / Math.max(1, textH * 0.72));
		const easedReveal = 1 - (1 - reveal) ** 2;
		text.style.transform = `translate3d(0, ${-window.scrollY}px, 0)`;
		img.style.transform = `scale(${(1.12 - p * 0.12).toFixed(3)})`;
		updateHeroPoints(1 - (1 - p) ** 2);
		carrier.style.transform = `translate3d(${(carrierStartX * (1 - easedReveal)).toFixed(1)}px, ${(carrierStartY * (1 - easedReveal)).toFixed(1)}px, 0) rotate(${((1 - easedReveal) * 5).toFixed(2)}deg)`;
		originalOwl.style.opacity = (1 - easedReveal).toFixed(3);
		pointingOwl.style.opacity = easedReveal.toFixed(3);
		paper.style.opacity = easedReveal.toFixed(3);
		dvr.style.opacity = easedReveal.toFixed(3);
		dvr.style.transform = `scale(${(0.96 + easedReveal * 0.04).toFixed(3)})`;
	}

	measureCarrierStart();

	if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
		carrier.style.transform = "none";
		originalOwl.style.opacity = "0";
		pointingOwl.style.opacity = "1";
		paper.style.opacity = "1";
		dvr.style.opacity = "1";
		dvr.style.transform = "none";
		updateHeroPoints(1);
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
