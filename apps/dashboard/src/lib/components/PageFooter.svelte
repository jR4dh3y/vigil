<script lang="ts">
	import type { Snippet } from "svelte";
	import { getPageFooter } from "$lib/nav/page-footer.svelte";

	let { children }: { children: Snippet } = $props();
	const footer = getPageFooter();

	// Register page content in the shell footer, outside the scrolling main area.
	$effect(() => {
		if (!footer) return;
		const content = children;
		footer.content = content;
		return () => {
			if (footer.content === content) {
				footer.content = undefined;
			}
		};
	});
</script>
