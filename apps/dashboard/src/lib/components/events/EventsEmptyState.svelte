<script lang="ts">
	import { Zap } from "lucide-svelte";

	type Props = {
		unacknowledgedOnly?: boolean;
	};

	let { unacknowledgedOnly = false }: Props = $props();
</script>

<div
	class="flex min-h-[320px] flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-zinc-800 bg-zinc-900/40 px-6 py-16 text-center"
>
	<span
		class="flex size-12 items-center justify-center rounded-xl bg-zinc-800/80 text-zinc-400"
	>
		<Zap class="size-6" />
	</span>
	<div class="flex flex-col gap-1">
		<p class="text-sm font-medium text-zinc-300">
			{unacknowledgedOnly ? "No unacknowledged events" : "No events yet"}
		</p>
		<p class="max-w-sm text-sm text-zinc-500">
			{unacknowledgedOnly
				? "Everything has been acknowledged. Clear the filter to see the full history."
				: "System and camera events will show up here as they occur."}
		</p>
	</div>
</div>
