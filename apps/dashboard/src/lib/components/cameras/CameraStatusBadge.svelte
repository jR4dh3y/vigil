<script lang="ts">
	import type { CameraStatus } from "$lib/cameras";
	import { statusLabel } from "$lib/cameras";

	type Props = {
		status: CameraStatus;
	};

	let { status }: Props = $props();

	const styles = $derived.by(() => {
		switch (status) {
			case "online":
				return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300";
			case "offline":
				return "border-red-500/30 bg-red-500/10 text-red-300";
			default:
				return "border-zinc-600/40 bg-zinc-800/80 text-zinc-400";
		}
	});

	const dot = $derived.by(() => {
		switch (status) {
			case "online":
				return "bg-emerald-400";
			case "offline":
				return "bg-red-400";
			default:
				return "bg-zinc-500";
		}
	});
</script>

<span
	class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs font-medium {styles}"
>
	<span class="size-1.5 rounded-full {dot}" aria-hidden="true"></span>
	{statusLabel(status)}
</span>
