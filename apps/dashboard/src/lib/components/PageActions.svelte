<script lang="ts">
	import type { Snippet } from "svelte";
	import {
		type PageActionsSide,
		getPageActionsHost,
	} from "$lib/nav/page-actions.svelte";

	type Props = {
		/** `start` = left of top bar (after title); `end` = right side. */
		side?: PageActionsSide;
		children: Snippet;
	};

	let { side = "end", children }: Props = $props();

	let portalEl = $state<HTMLDivElement | null>(null);

	/**
	 * Portal action controls into the SubNav host so they appear in the top bar.
	 * Uses a placeholder so Svelte can reclaim the node on unmount cleanly.
	 */
	$effect(() => {
		const node = portalEl;
		if (!node) {
			return;
		}

		const host = getPageActionsHost(side);
		if (!host) {
			return;
		}

		const parent = node.parentNode;
		if (!parent) {
			return;
		}

		const placeholder = document.createComment("page-actions");
		parent.replaceChild(placeholder, node);
		host.appendChild(node);

		return () => {
			if (node.parentNode === host) {
				host.removeChild(node);
			}
			if (placeholder.parentNode) {
				placeholder.parentNode.replaceChild(node, placeholder);
			}
		};
	});

	const alignClass = $derived(
		side === "start"
			? "flex min-w-0 items-center gap-2"
			: "flex min-w-0 flex-wrap items-center justify-end gap-2",
	);
</script>

<!-- Keep a hidden home slot; the inner node is moved into the top bar. -->
<div class="hidden" aria-hidden="true">
	<div bind:this={portalEl} class={alignClass}>
		{@render children()}
	</div>
</div>
