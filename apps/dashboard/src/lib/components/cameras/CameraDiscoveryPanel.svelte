<script lang="ts">
	import { KeyRound, Plus, Radar, Search, ShieldCheck } from "lucide-svelte";
	import type { DiscoveredCamera } from "$lib/cameras";

	type Props = {
		cameras: DiscoveredCamera[];
		scanning?: boolean;
		hasScanned?: boolean;
		error?: string | null;
		onScan: (username: string, password: string) => void | Promise<void>;
		onSelect: (camera: DiscoveredCamera) => void;
		onAddSelected: (cameras: DiscoveredCamera[]) => void | Promise<void>;
		onManual: () => void;
		addingSelected?: boolean;
		addProgress?: { completed: number; total: number } | null;
	};

	let {
		cameras,
		scanning = false,
		hasScanned = false,
		error = null,
		onScan,
		onSelect,
		onAddSelected,
		onManual,
		addingSelected = false,
		addProgress = null,
	}: Props = $props();

	let username = $state("");
	let password = $state("");
	let validationError = $state<string | null>(null);
	let selectedIds = $state<string[]>([]);
	let selectedCameras = $derived(cameras.filter((camera) => selectedIds.includes(camera.id)));
	let allSelected = $derived(cameras.length > 0 && selectedCameras.length === cameras.length);
	let busy = $derived(scanning || addingSelected);

	const inputClass =
		"rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50";

	async function handleScan(event: Event) {
		event.preventDefault();
		validationError = null;
		if (!username.trim() || !password) {
			validationError = "Enter the camera or NVR username/ID and password before scanning.";
			return;
		}
		selectedIds = [];
		await onScan(username.trim(), password);
	}

	function toggleSelected(camera: DiscoveredCamera) {
		if (selectedIds.includes(camera.id)) {
			selectedIds = selectedIds.filter((id) => id !== camera.id);
			return;
		}
		selectedIds = [...selectedIds, camera.id];
	}

	function toggleAll() {
		selectedIds = allSelected ? [] : cameras.map((camera) => camera.id);
	}

	async function handleAddSelected() {
		if (selectedCameras.length === 0 || addingSelected) {
			return;
		}
		try {
			await onAddSelected(selectedCameras);
			selectedIds = [];
		} catch {
			// The parent displays the detailed bulk-create error and keeps the selection.
		}
	}
</script>

<div class="flex flex-col gap-5">
	<div>
		<div>
			<div class="mb-2 flex items-center gap-2 text-emerald-400">
				<Search class="size-4" />
				<span class="text-xs font-semibold uppercase tracking-[0.18em]">Network discovery</span>
			</div>
			<h1 class="text-xl font-semibold text-zinc-100">Find a camera on your network</h1>
			<p class="mt-1 max-w-xl text-sm text-zinc-400">
				Enter the camera or NVR username/ID and password, then scan. The NVR uses those credentials
				to validate ONVIF devices and Dahua RTSP channels such as
				<code class="font-mono text-zinc-300">/cam/realmonitor</code>.
			</p>
		</div>
	</div>

	<form class="rounded-xl border border-zinc-800 bg-zinc-950/40 p-4" onsubmit={handleScan} novalidate>
		<div class="mb-3 flex items-center gap-2 text-zinc-300">
			<KeyRound class="size-4 text-emerald-400" />
			<span class="text-sm font-medium">Camera / NVR credentials</span>
		</div>
		<div class="grid gap-4 sm:grid-cols-2">
			<div class="flex flex-col gap-1.5">
				<label for="discovery-username" class="text-sm font-medium text-zinc-300">
					Username / ID
				</label>
				<input
					id="discovery-username"
					type="text"
					autocomplete="username"
					bind:value={username}
					disabled={busy}
					class={inputClass}
					placeholder="admin"
				/>
			</div>
			<div class="flex flex-col gap-1.5">
				<label for="discovery-password" class="text-sm font-medium text-zinc-300">Password</label>
				<input
					id="discovery-password"
					type="password"
					autocomplete="current-password"
					bind:value={password}
					disabled={busy}
					class={inputClass}
					placeholder="Camera password"
				/>
			</div>
		</div>
		{#if validationError}
			<p class="mt-3 text-sm text-red-400" role="alert">{validationError}</p>
		{/if}
		<button
			type="submit"
			disabled={busy}
			class="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
		>
			<Radar class="size-4 {scanning ? 'animate-spin' : ''}" />
			{scanning ? "Scanning…" : hasScanned ? "Scan again" : "Scan network"}
		</button>
	</form>

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
				<p class="text-sm font-medium text-zinc-200">Looking for RTSP/ONVIF cameras…</p>
				<p class="text-xs text-zinc-500">This usually takes a few seconds on a local network.</p>
			</div>
		</div>
	{:else if cameras.length > 0}
		<div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-zinc-800 bg-zinc-950/30 px-3 py-2">
			<label class="inline-flex items-center gap-2 text-sm text-zinc-300">
				<input
					type="checkbox"
					checked={allSelected}
					disabled={busy}
					onchange={toggleAll}
					class="size-4 rounded border-zinc-600 bg-zinc-900 text-emerald-500 focus:ring-emerald-500/40"
				/>
				Select all
			</label>
			<span class="text-xs text-zinc-500">
				{selectedCameras.length} of {cameras.length} selected
			</span>
		</div>
		<div class="grid gap-3">
			{#each cameras as camera (camera.id)}
				<div
					class="flex items-center justify-between gap-3 rounded-xl border border-zinc-800 bg-zinc-950/60 px-3 py-3 transition-colors hover:border-emerald-500/40 hover:bg-zinc-900"
				>
					<label class="flex min-w-0 flex-1 cursor-pointer items-center gap-3">
						<input
							type="checkbox"
							checked={selectedIds.includes(camera.id)}
							disabled={busy}
							onchange={() => toggleSelected(camera)}
							aria-label={`Select ${camera.name}`}
							class="size-4 shrink-0 rounded border-zinc-600 bg-zinc-900 text-emerald-500 focus:ring-emerald-500/40"
						/>
						<span class="min-w-0">
							<span class="block truncate text-sm font-medium text-zinc-100">{camera.name}</span>
							<span class="mt-1 block truncate font-mono text-xs text-zinc-500">{camera.host}</span>
						</span>
					</label>
					<button
						type="button"
						disabled={busy}
						onclick={() => onSelect(camera)}
						class="shrink-0 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm font-medium text-emerald-400 transition-colors hover:border-emerald-500/50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
					>
						Configure
					</button>
				</div>
			{/each}
		</div>
		{#if selectedCameras.length > 0}
			<button
				type="button"
				disabled={busy}
				onclick={handleAddSelected}
				class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
			>
				<Plus class="size-4" />
				{#if addingSelected && addProgress}
					Adding {addProgress.completed}/{addProgress.total}…
				{:else}
					Add {selectedCameras.length} selected camera{selectedCameras.length === 1 ? "" : "s"}
				{/if}
			</button>
		{/if}
	{:else if hasScanned && !error}
		<div class="rounded-xl border border-dashed border-zinc-800 bg-zinc-950/40 px-4 py-6 text-center">
			<p class="text-sm text-zinc-300">No RTSP/ONVIF cameras found.</p>
			<p class="mt-1 text-xs text-zinc-500">
				Check the credentials and make sure the camera and NVR are on the same network, or
				continue with manual setup.
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
			disabled={busy}
			onclick={onManual}
			class="rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-600 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
		>
			Enter camera manually
		</button>
	</div>
</div>
