<script lang="ts">
	import { untrack } from "svelte";
	import { Radar } from "lucide-svelte";
	import type { ProbeResult } from "$lib/cameras";
	import {
		createCameraFormSchema,
		type CreateCameraFormValues,
		type EditCameraFormValues,
		editCameraFormSchema,
		fieldErrorsFromZod,
		resolveProbeRtspUrl,
	} from "$lib/cameras";
	import ProbeResultPanel from "./ProbeResultPanel.svelte";

	type Mode = "create" | "edit";

	type Props = {
		mode: Mode;
		initial?: Partial<CreateCameraFormValues>;
		submitting?: boolean;
		probing?: boolean;
		serverError?: string | null;
		probeResult?: ProbeResult | null;
		probeError?: string | null;
		onSubmit: (values: CreateCameraFormValues | EditCameraFormValues) => void | Promise<void>;
		onProbe: (input: {
			rtspUrl: string;
			username?: string;
			password?: string;
		}) => void | Promise<void>;
		onDelete?: () => void | Promise<void>;
		deleting?: boolean;
	};

	let {
		mode,
		initial = {},
		submitting = false,
		probing = false,
		serverError = null,
		probeResult = null,
		probeError = null,
		onSubmit,
		onProbe,
		onDelete,
		deleting = false,
	}: Props = $props();

	// Seed local form state once from props; parent remounts via {#key} after saves.
	const seed = untrack(() => ({
		name: initial.name ?? "",
		host: initial.host ?? "",
		username: initial.username ?? "",
		password: initial.password ?? "",
		enabled: initial.enabled ?? true,
		liveRtspUrl: initial.liveRtspUrl ?? "",
		recordRtspUrl: initial.recordRtspUrl ?? "",
	}));

	let name = $state(seed.name);
	let host = $state(seed.host);
	let username = $state(seed.username);
	let password = $state(seed.password);
	let enabled = $state(seed.enabled);
	let liveRtspUrl = $state(seed.liveRtspUrl);
	let recordRtspUrl = $state(seed.recordRtspUrl);
	let fieldErrors = $state<Record<string, string>>({});
	let localProbeError = $state<string | null>(null);

	const busy = $derived(submitting || probing || deleting);

	const inputClass =
		"rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50";

	function currentValues(): CreateCameraFormValues {
		return {
			name,
			host,
			username,
			password,
			enabled,
			liveRtspUrl,
			recordRtspUrl,
		};
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();
		fieldErrors = {};
		localProbeError = null;

		const schema = mode === "create" ? createCameraFormSchema : editCameraFormSchema;
		const parsed = schema.safeParse(currentValues());
		if (!parsed.success) {
			fieldErrors = fieldErrorsFromZod(parsed.error);
			return;
		}

		await onSubmit(parsed.data);
	}

	async function handleProbe() {
		localProbeError = null;
		fieldErrors = {};

		const rtspUrl = resolveProbeRtspUrl(liveRtspUrl, host);
		if (!rtspUrl) {
			localProbeError =
				"Enter a live RTSP URL, or set host to an rtsp:// address, before probing.";
			return;
		}

		const payload: { rtspUrl: string; username?: string; password?: string } = { rtspUrl };
		const user = username.trim();
		if (user) {
			payload.username = user;
		}
		if (password) {
			payload.password = password;
		}

		await onProbe(payload);
	}
</script>

<form class="flex flex-col gap-6" onsubmit={handleSubmit} novalidate>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex flex-col gap-1.5 sm:col-span-2">
			<label for="camera-name" class="text-sm font-medium text-zinc-300">Name</label>
			<input
				id="camera-name"
				name="name"
				type="text"
				bind:value={name}
				disabled={busy}
				class={inputClass}
				placeholder="Front door"
			/>
			{#if fieldErrors.name}
				<p class="text-xs text-red-400">{fieldErrors.name}</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5 sm:col-span-2">
			<label for="camera-host" class="text-sm font-medium text-zinc-300">Host</label>
			<input
				id="camera-host"
				name="host"
				type="text"
				bind:value={host}
				disabled={busy}
				class="{inputClass} font-mono"
				placeholder="192.168.1.50 or rtsp://..."
			/>
			{#if fieldErrors.host}
				<p class="text-xs text-red-400">{fieldErrors.host}</p>
			{:else}
				<p class="text-xs text-zinc-500">Display host or primary RTSP base for the camera.</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5">
			<label for="camera-username" class="text-sm font-medium text-zinc-300">Username</label>
			<input
				id="camera-username"
				name="username"
				type="text"
				autocomplete="off"
				bind:value={username}
				disabled={busy}
				class={inputClass}
				placeholder="Optional"
			/>
			{#if fieldErrors.username}
				<p class="text-xs text-red-400">{fieldErrors.username}</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5">
			<label for="camera-password" class="text-sm font-medium text-zinc-300">Password</label>
			<input
				id="camera-password"
				name="password"
				type="password"
				autocomplete="new-password"
				bind:value={password}
				disabled={busy}
				class={inputClass}
				placeholder={mode === "edit" ? "Leave blank to keep current" : "Optional"}
			/>
			{#if fieldErrors.password}
				<p class="text-xs text-red-400">{fieldErrors.password}</p>
			{:else if mode === "edit"}
				<p class="text-xs text-zinc-500">Leave empty to keep the existing password.</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5 sm:col-span-2">
			<label for="camera-live-rtsp" class="text-sm font-medium text-zinc-300"
				>Live RTSP URL</label
			>
			<input
				id="camera-live-rtsp"
				name="liveRtspUrl"
				type="url"
				bind:value={liveRtspUrl}
				disabled={busy}
				class="{inputClass} font-mono"
				placeholder="rtsp://192.168.1.50:554/stream1"
			/>
			{#if fieldErrors.liveRtspUrl}
				<p class="text-xs text-red-400">{fieldErrors.liveRtspUrl}</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5 sm:col-span-2">
			<label for="camera-record-rtsp" class="text-sm font-medium text-zinc-300"
				>Record RTSP URL</label
			>
			<input
				id="camera-record-rtsp"
				name="recordRtspUrl"
				type="url"
				bind:value={recordRtspUrl}
				disabled={busy}
				class="{inputClass} font-mono"
				placeholder="rtsp://192.168.1.50:554/stream0 (optional)"
			/>
			{#if fieldErrors.recordRtspUrl}
				<p class="text-xs text-red-400">{fieldErrors.recordRtspUrl}</p>
			{/if}
		</div>

		<div class="flex items-center gap-3 sm:col-span-2">
			<button
				type="button"
				role="switch"
				aria-checked={enabled}
				aria-label="Enabled"
				disabled={busy}
				onclick={() => {
					enabled = !enabled;
				}}
				class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border transition-colors disabled:cursor-not-allowed disabled:opacity-50
					{enabled
					? 'border-emerald-500/40 bg-emerald-600'
					: 'border-zinc-700 bg-zinc-800'}"
			>
				<span
					class="pointer-events-none absolute top-0.5 left-0.5 size-5 rounded-full bg-white shadow transition-transform
						{enabled ? 'translate-x-5' : 'translate-x-0'}"
				></span>
			</button>
			<div class="flex flex-col">
				<span class="text-sm font-medium text-zinc-300">Enabled</span>
				<span class="text-xs text-zinc-500">
					{enabled ? "Camera will be active for live and recording." : "Camera is disabled."}
				</span>
			</div>
		</div>
	</div>

	<div class="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<p class="text-sm font-medium text-zinc-200">Probe stream</p>
				<p class="text-xs text-zinc-500">
					Test connectivity using the live RTSP URL (or host if it is an RTSP URL).
				</p>
			</div>
			<button
				type="button"
				disabled={busy}
				onclick={handleProbe}
				class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 transition-colors hover:border-zinc-600 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
			>
				<Radar class="size-3.5 {probing ? 'animate-spin' : ''}" />
				{probing ? "Probing…" : "Probe"}
			</button>
		</div>

		{#if localProbeError || probeError}
			<p
				class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
				role="alert"
			>
				{localProbeError ?? probeError}
			</p>
		{/if}

		{#if probeResult}
			<ProbeResultPanel result={probeResult} />
		{/if}
	</div>

	{#if serverError}
		<p
			class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{serverError}
		</p>
	{/if}

	<div class="flex flex-wrap items-center justify-between gap-3 border-t border-zinc-800 pt-4">
		<div>
			{#if mode === "edit" && onDelete}
				<button
					type="button"
					disabled={busy}
					onclick={() => onDelete()}
					class="rounded-lg border border-red-500/30 bg-red-500/10 px-3.5 py-2 text-sm font-medium text-red-300 transition-colors hover:border-red-500/50 hover:bg-red-500/15 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{deleting ? "Deleting…" : "Delete camera"}
				</button>
			{/if}
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<a
				href="/cameras"
				class="rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2 text-sm text-zinc-300 no-underline transition-colors hover:border-zinc-600 hover:bg-zinc-800 hover:text-zinc-100"
			>
				Cancel
			</a>
			<button
				type="submit"
				disabled={busy}
				class="inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
			>
				{#if submitting}
					{mode === "create" ? "Creating…" : "Saving…"}
				{:else}
					{mode === "create" ? "Create camera" : "Save changes"}
				{/if}
			</button>
		</div>
	</div>
</form>
