<script lang="ts">
	import { untrack } from "svelte";
	import { Check, Database, Save, Server } from "lucide-svelte";
	import { fieldErrorsFromZod, type SettingsFormValues, settingsFormSchema } from "$lib/system";

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

<form class="flex flex-col" onsubmit={handleSubmit} novalidate>
	<section class="grid gap-5 border-b border-zinc-800 pb-6 sm:grid-cols-[180px_minmax(0,1fr)]">
		<div>
			<div
				class="mb-2 flex size-8 items-center justify-center rounded-lg bg-zinc-800/80 text-zinc-400"
			>
				<Server class="size-4" />
			</div>
			<h3 class="text-sm font-medium text-zinc-200">Identity</h3>
			<p class="mt-1 text-xs leading-5 text-zinc-500">Shown throughout the dashboard.</p>
		</div>
		<div class="flex flex-col gap-1.5">
			<label for="settings-site-name" class="text-xs font-medium text-zinc-400">Site name</label>
			<input
				id="settings-site-name"
				name="siteName"
				type="text"
				bind:value={siteName}
				disabled={readonly || submitting}
				class="h-10 rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 text-sm text-zinc-100 outline-none ring-emerald-500/30 transition placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
				placeholder="Home NVR"
			/>
			{#if fieldErrors.siteName}
				<p class="text-xs text-red-400">{fieldErrors.siteName}</p>
			{/if}
		</div>
	</section>

	<section class="grid gap-5 pt-6 sm:grid-cols-[180px_minmax(0,1fr)]">
		<div>
			<div
				class="mb-2 flex size-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400"
			>
				<Database class="size-4" />
			</div>
			<h3 class="text-sm font-medium text-zinc-200">Local recording</h3>
			<p class="mt-1 text-xs leading-5 text-zinc-500">
				Control where footage is stored and for how long.
			</p>
		</div>

		<div class="flex min-w-0 flex-col gap-5">
			<label
				class="flex items-center justify-between gap-4 rounded-xl border border-zinc-800 bg-zinc-950/55 p-3.5
					{readonly || submitting ? 'cursor-not-allowed opacity-60' : 'cursor-pointer hover:border-zinc-700'}"
			>
				<span class="min-w-0">
					<span class="block text-sm font-medium text-zinc-200">Continuous recording</span>
					<span class="mt-0.5 block text-xs leading-5 text-zinc-500">
						Live view remains available when this is off.
					</span>
				</span>
				<span class="relative inline-flex shrink-0">
					<input
						type="checkbox"
						class="peer sr-only"
						bind:checked={recordingEnabled}
						disabled={readonly || submitting}
					/>
					<span
						class="h-6 w-11 rounded-full border border-zinc-700 bg-zinc-800 transition-colors peer-checked:border-emerald-500 peer-checked:bg-emerald-500 peer-focus-visible:ring-2 peer-focus-visible:ring-emerald-500/30 peer-disabled:cursor-not-allowed"
					></span>
					<span
						class="pointer-events-none absolute top-1 left-1 size-4 rounded-full bg-zinc-400 shadow-sm transition-transform peer-checked:translate-x-5 peer-checked:bg-white"
					></span>
				</span>
			</label>

			<div class="flex flex-col gap-1.5">
				<label for="settings-recordings-dir" class="text-xs font-medium text-zinc-400">
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
					class="h-10 min-w-0 rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 font-mono text-sm text-zinc-100 outline-none ring-emerald-500/30 transition placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
					placeholder="/var/lib/nvr/recordings"
				/>
				<p class="text-xs leading-5 text-zinc-500">
					A path on the recorder host, such as
					<code class="rounded bg-zinc-800/80 px-1 py-0.5 text-zinc-400">/mnt/storage/nvr</code>.
				</p>
				{#if fieldErrors.recordingsDir}
					<p class="text-xs text-red-400">{fieldErrors.recordingsDir}</p>
				{/if}
				{#if fieldErrors.recordingEnabled}
					<p class="text-xs text-red-400">{fieldErrors.recordingEnabled}</p>
				{/if}
			</div>

			<div class="flex flex-col gap-1.5">
				<label for="settings-retention" class="text-xs font-medium text-zinc-400">
					Retention period
				</label>
				<div class="relative max-w-48">
					<input
						id="settings-retention"
						name="retentionDays"
						type="number"
						min="1"
						max="3650"
						step="1"
						bind:value={retentionDays}
						disabled={readonly || submitting}
						class="h-10 w-full rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 pr-14 text-sm text-zinc-100 outline-none ring-emerald-500/30 transition focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
					/>
					<span
						class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-zinc-500"
						>days</span
					>
				</div>
				<p class="text-xs leading-5 text-zinc-500">
					Older recordings are eligible for automatic deletion.
				</p>
				{#if fieldErrors.retentionDays}
					<p class="text-xs text-red-400">{fieldErrors.retentionDays}</p>
				{/if}
			</div>
		</div>
	</section>

	{#if serverError}
		<p
			class="mt-6 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{serverError}
		</p>
	{/if}

	{#if successMessage}
		<p
			class="mt-6 flex items-center gap-2 rounded-lg border border-emerald-500/25 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300"
			role="status"
		>
			<Check class="size-4 shrink-0" />
			{successMessage}
		</p>
	{/if}

	{#if readonly}
		<p class="mt-6 border-t border-zinc-800 pt-5 text-xs text-zinc-500">
			Only administrators can change recorder settings.
		</p>
	{:else}
		<div class="mt-6 flex items-center justify-end border-t border-zinc-800 pt-5">
			<button
				type="submit"
				disabled={submitting}
				class="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-500 px-4 py-2.5 text-sm font-medium text-zinc-950 shadow-sm shadow-emerald-950/20 transition-colors hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60"
			>
				<Save class="size-4" />
				{submitting ? "Saving…" : "Save changes"}
			</button>
		</div>
	{/if}
</form>
