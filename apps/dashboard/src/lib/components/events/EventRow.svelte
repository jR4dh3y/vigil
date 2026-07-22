<script lang="ts">
	import { Check, Camera } from "lucide-svelte";
	import type { Event } from "$lib/events";
	import { formatEventTime } from "$lib/events";
	import EventSeverityBadge from "./EventSeverityBadge.svelte";

	type Props = {
		event: Event;
		acknowledging?: boolean;
		onAcknowledge?: (id: string) => void | Promise<void>;
	};

	let { event, acknowledging = false, onAcknowledge }: Props = $props();
</script>

<article
	class="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 transition-colors sm:flex-row sm:items-start sm:justify-between
		{event.acknowledged ? 'opacity-80' : 'hover:border-zinc-700'}"
>
	<div class="min-w-0 flex-1 flex flex-col gap-2">
		<div class="flex flex-wrap items-center gap-2">
			<EventSeverityBadge severity={event.severity} />
			<span
				class="rounded-md border border-zinc-700/80 bg-zinc-800/80 px-2 py-0.5 font-mono text-xs text-zinc-400"
			>
				{event.type}
			</span>
			{#if event.acknowledged}
				<span
					class="rounded-md border border-zinc-700/60 bg-zinc-800/60 px-2 py-0.5 text-xs text-zinc-500"
				>
					Acknowledged
				</span>
			{/if}
		</div>

		<div class="flex flex-col gap-0.5">
			<h2 class="text-sm font-semibold text-zinc-100">{event.title}</h2>
			{#if event.message}
				<p class="text-sm text-zinc-400">{event.message}</p>
			{/if}
		</div>

		<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-zinc-500">
			<time datetime={event.startedAt}>{formatEventTime(event.startedAt)}</time>
			{#if event.cameraId}
				<a
					href="/cameras/{event.cameraId}"
					class="inline-flex items-center gap-1 text-emerald-400 no-underline hover:text-emerald-300"
				>
					<Camera class="size-3" />
					View camera
				</a>
			{/if}
		</div>
	</div>

	{#if !event.acknowledged && onAcknowledge}
		<button
			type="button"
			class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 transition-colors hover:border-emerald-500/40 hover:bg-emerald-500/10 hover:text-emerald-300 disabled:cursor-not-allowed disabled:opacity-50"
			disabled={acknowledging}
			onclick={() => onAcknowledge(event.id)}
		>
			<Check class="size-3.5" />
			{acknowledging ? "Acknowledging…" : "Acknowledge"}
		</button>
	{/if}
</article>
