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

/* Hero scroll reveal: the paper text zone lifts off the sticky night-sky
   image while the image settles from a close crop on the owl to the full
   scene (sky, house, fence), then the next section slides over it. */
const heroText = document.querySelector<HTMLElement>("[data-hero-text]");
const heroImg = document.querySelector<HTMLElement>("[data-hero-img]");

if (heroText && heroImg) {
	function updateHero(text: HTMLElement, img: HTMLElement) {
		const textH = text.offsetHeight;
		const p = Math.min(1, Math.max(0, window.scrollY / Math.max(1, textH)));
		text.style.transform = `translate3d(0, ${-window.scrollY}px, 0)`;
		img.style.transform = `scale(${(1.12 - p * 0.12).toFixed(3)})`;
	}

	if (!window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
		updateHero(heroText, heroImg);
		let ticking = false;
		window.addEventListener(
			"scroll",
			() => {
				if (ticking) return;
				ticking = true;
				requestAnimationFrame(() => {
					updateHero(heroText, heroImg);
					ticking = false;
				});
			},
			{ passive: true },
		);
		window.addEventListener("resize", () => updateHero(heroText, heroImg));
	}
}
