<script lang="ts">
	import { AlertTriangle, CheckCircle2, XCircle } from "lucide-svelte";
	import type { ProbeResult } from "$lib/cameras";

	type Props = {
		result: ProbeResult;
	};

	let { result }: Props = $props();

	const resolution = $derived(
		result.width && result.height ? `${result.width}×${result.height}` : null,
	);
</script>

<div
	class="rounded-lg border px-3 py-3 text-sm
		{result.reachable
		? 'border-emerald-500/25 bg-emerald-500/5'
		: 'border-red-500/25 bg-red-500/5'}"
	role="status"
>
	<div class="flex items-start gap-2">
		{#if result.reachable}
			<CheckCircle2 class="mt-0.5 size-4 shrink-0 text-emerald-400" />
		{:else}
			<XCircle class="mt-0.5 size-4 shrink-0 text-red-400" />
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<p class="font-medium {result.reachable ? 'text-emerald-200' : 'text-red-200'}">
				{result.reachable ? "Stream reachable" : "Stream not reachable"}
			</p>
			{#if result.error}
				<p class="text-xs text-red-300/90">{result.error}</p>
			{/if}
			{#if result.reachable}
				<ul class="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-zinc-400">
					{#if result.codec}
						<li>
							<span class="text-zinc-500">Codec</span>
							<span class="ml-1 text-zinc-200">{result.codec}</span>
						</li>
					{/if}
					{#if resolution}
						<li>
							<span class="text-zinc-500">Resolution</span>
							<span class="ml-1 text-zinc-200">{resolution}</span>
						</li>
					{/if}
				</ul>
			{/if}
			{#if result.h265}
				<p
					class="mt-2 flex items-start gap-1.5 rounded-md border border-amber-500/30 bg-amber-500/10 px-2.5 py-2 text-xs text-amber-200"
				>
					<AlertTriangle class="mt-0.5 size-3.5 shrink-0 text-amber-400" />
					<span>
						H.265 / HEVC detected. Browser playback may be limited; prefer H.264 for live
						view where possible.
					</span>
				</p>
			{/if}
		</div>
	</div>
</div>
