<script lang="ts">
	import { Radar, Search, ShieldCheck } from "lucide-svelte";
	import type { DiscoveredCamera } from "$lib/cameras";

	type Props = {
		cameras: DiscoveredCamera[];
		scanning?: boolean;
		error?: string | null;
		onScan: () => void | Promise<void>;
		onSelect: (camera: DiscoveredCamera) => void;
		onManual: () => void;
	};

	let {
		cameras,
		scanning = false,
		error = null,
		onScan,
		onSelect,
		onManual,
	}: Props = $props();
</script>

<div class="flex flex-col gap-5">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<div class="mb-2 flex items-center gap-2 text-emerald-400">
				<Search class="size-4" />
				<span class="text-xs font-semibold uppercase tracking-[0.18em]">Network discovery</span>
			</div>
			<h1 class="text-xl font-semibold text-zinc-100">Find a camera on your network</h1>
			<p class="mt-1 max-w-xl text-sm text-zinc-400">
				We scan for ONVIF cameras before asking for credentials. Your username and password stay on
				the NVR and are only requested after you choose a camera.
			</p>
		</div>
		<button
			type="button"
			disabled={scanning}
			onclick={() => onScan()}
			class="inline-flex items-center gap-2 rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-3.5 py-2 text-sm font-medium text-emerald-300 transition-colors hover:border-emerald-400/60 hover:bg-emerald-500/15 disabled:cursor-not-allowed disabled:opacity-60"
		>
			<Radar class="size-4 {scanning ? 'animate-spin' : ''}" />
			{scanning ? "Scanning…" : "Scan again"}
		</button>
	</div>

	{#if error}
		<p
			class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{error}
		</p>
	{/if}

	{#if scanning}
		<div
			class="flex items-center gap-3 rounded-xl border border-zinc-800 bg-zinc-950/60 px-4 py-5"
			role="status"
			aria-live="polite"
		>
			<Radar class="size-5 animate-spin text-emerald-400" />
			<div>
				<p class="text-sm font-medium text-zinc-200">Looking for ONVIF cameras…</p>
				<p class="text-xs text-zinc-500">This usually takes a few seconds on a local network.</p>
			</div>
		</div>
	{:else if cameras.length > 0}
		<div class="grid gap-3">
			{#each cameras as camera (camera.id)}
				<button
					type="button"
					class="flex items-center justify-between gap-4 rounded-xl border border-zinc-800 bg-zinc-950/60 px-4 py-4 text-left transition-colors hover:border-emerald-500/40 hover:bg-zinc-900"
					onclick={() => onSelect(camera)}
				>
					<span class="min-w-0">
						<span class="block truncate text-sm font-medium text-zinc-100">{camera.name}</span>
						<span class="mt-1 block truncate font-mono text-xs text-zinc-500">{camera.host}</span>
					</span>
					<span class="shrink-0 text-sm font-medium text-emerald-400">Use camera</span>
				</button>
			{/each}
		</div>
	{:else if !error}
		<div class="rounded-xl border border-dashed border-zinc-800 bg-zinc-950/40 px-4 py-6 text-center">
			<p class="text-sm text-zinc-300">No ONVIF cameras found.</p>
			<p class="mt-1 text-xs text-zinc-500">
				Make sure the camera and NVR are on the same network, or continue with manual setup.
			</p>
		</div>
	{/if}

	<div class="flex flex-wrap items-center justify-between gap-3 border-t border-zinc-800 pt-4">
		<div class="flex items-center gap-2 text-xs text-zinc-500">
			<ShieldCheck class="size-4 text-emerald-500" />
			<span>Discovery does not store credentials.</span>
		</div>
		<button
			type="button"
			disabled={scanning}
			onclick={onManual}
			class="rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-600 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
		>
			Enter camera manually
		</button>
	</div>
</div>
