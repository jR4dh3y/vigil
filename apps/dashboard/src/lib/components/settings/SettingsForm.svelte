<script lang="ts">
	import { untrack } from "svelte";
	import {
		fieldErrorsFromZod,
		type SettingsFormValues,
		settingsFormSchema,
	} from "$lib/system";

	type Props = {
		initial: SettingsFormValues;
		readonly?: boolean;
		submitting?: boolean;
		serverError?: string | null;
		successMessage?: string | null;
		onSubmit: (values: SettingsFormValues) => void | Promise<void>;
	};

	let {
		initial,
		readonly = false,
		submitting = false,
		serverError = null,
		successMessage = null,
		onSubmit,
	}: Props = $props();

	// Seed once; parent remounts via {#key} when server values change.
	const seed = untrack(() => ({
		siteName: initial.siteName,
		retentionDays: String(initial.retentionDays),
		recordingsDir: initial.recordingsDir,
		recordingEnabled: initial.recordingEnabled,
	}));
	let siteName = $state(seed.siteName);
	let retentionDays = $state(seed.retentionDays);
	let recordingsDir = $state(seed.recordingsDir);
	let recordingEnabled = $state(seed.recordingEnabled);
	let fieldErrors = $state<Record<string, string>>({});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		if (readonly) {
			return;
		}
		fieldErrors = {};

		const parsed = settingsFormSchema.safeParse({
			siteName,
			retentionDays,
			recordingsDir,
			recordingEnabled,
		});
		if (!parsed.success) {
			fieldErrors = fieldErrorsFromZod(parsed.error);
			return;
		}

		await onSubmit(parsed.data);
	}
</script>

<form class="flex flex-col gap-5" onsubmit={handleSubmit} novalidate>
	<div class="flex flex-col gap-1.5">
		<label for="settings-site-name" class="text-sm font-medium text-zinc-300">Site name</label>
		<input
			id="settings-site-name"
			name="siteName"
			type="text"
			bind:value={siteName}
			disabled={readonly || submitting}
			class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
			placeholder="Home NVR"
		/>
		{#if fieldErrors.siteName}
			<p class="text-xs text-red-400">{fieldErrors.siteName}</p>
		{/if}
	</div>

	<div class="flex flex-col gap-3 rounded-lg border border-zinc-800 bg-zinc-950/50 p-4">
		<div class="flex flex-col gap-1">
			<p class="text-sm font-medium text-zinc-200">Recording</p>
			<p class="text-xs text-zinc-500">
				Choose a folder on this NVR host and turn continuous recording on or off.
			</p>
		</div>

		<label
			class="flex cursor-pointer items-center justify-between gap-3 rounded-lg border border-zinc-800 bg-zinc-900/60 px-3 py-2.5
				{readonly || submitting ? 'cursor-not-allowed opacity-60' : ''}"
		>
			<div class="flex flex-col gap-0.5">
				<span class="text-sm font-medium text-zinc-200">Recording enabled</span>
				<span class="text-xs text-zinc-500">
					When off, live view still works but new segments are not written.
				</span>
			</div>
			<input
				type="checkbox"
				class="size-4 rounded border-zinc-600 bg-zinc-950 text-emerald-500 focus:ring-emerald-500/40"
				bind:checked={recordingEnabled}
				disabled={readonly || submitting}
			/>
		</label>

		<div class="flex flex-col gap-1.5">
			<label for="settings-recordings-dir" class="text-sm font-medium text-zinc-300">
				Recording location
			</label>
			<input
				id="settings-recordings-dir"
				name="recordingsDir"
				type="text"
				bind:value={recordingsDir}
				disabled={readonly || submitting}
				spellcheck="false"
				autocomplete="off"
				class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
				placeholder="/var/lib/nvr/recordings"
			/>
			<p class="text-xs text-zinc-500">
				Path on the NVR server (not your browser). Absolute paths are recommended, e.g.
				<code class="text-zinc-400">/mnt/storage/nvr</code> or
				<code class="text-zinc-400">./recordings</code>.
			</p>
			{#if fieldErrors.recordingsDir}
				<p class="text-xs text-red-400">{fieldErrors.recordingsDir}</p>
			{/if}
			{#if fieldErrors.recordingEnabled}
				<p class="text-xs text-red-400">{fieldErrors.recordingEnabled}</p>
			{/if}
		</div>
	</div>

	<div class="flex flex-col gap-1.5">
		<label for="settings-retention" class="text-sm font-medium text-zinc-300"
			>Recording retention (days)</label
		>
		<input
			id="settings-retention"
			name="retentionDays"
			type="number"
			min="1"
			max="3650"
			step="1"
			bind:value={retentionDays}
			disabled={readonly || submitting}
			class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
		/>
		<p class="text-xs text-zinc-500">
			Recordings older than this many days may be deleted by retention jobs.
		</p>
		{#if fieldErrors.retentionDays}
			<p class="text-xs text-red-400">{fieldErrors.retentionDays}</p>
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

	{#if successMessage}
		<p
			class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300"
			role="status"
		>
			{successMessage}
		</p>
	{/if}

	{#if readonly}
		<p class="text-xs text-zinc-500">Only administrators can change settings.</p>
	{:else}
		<button
			type="submit"
			disabled={submitting}
			class="inline-flex w-fit items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
		>
			{submitting ? "Saving…" : "Save settings"}
		</button>
	{/if}
</form>
