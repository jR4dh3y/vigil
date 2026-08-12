<script lang="ts">
	import { Archive, ChevronDown, Cloud, CloudOff, Link2, Unplug } from "lucide-svelte";
	import type { GDriveStatus } from "$lib/storage";
	import GoogleDriveConfigurationForm from "./GoogleDriveConfigurationForm.svelte";
	import { driveConnectionBadgeClass, driveConnectionLabel, formatConnectedAt } from "$lib/storage";

	type Props = {
		status: GDriveStatus;
		readonly?: boolean;
		connecting?: boolean;
		disconnecting?: boolean;
		archiving?: boolean;
		configuring?: boolean;
		serverError?: string | null;
		onConnect: () => void | Promise<void>;
		onDisconnect: () => void | Promise<void>;
		onArchive: () => void | Promise<void>;
		onConfigure: (
			values: import("$lib/storage").GDriveConfigurationRequest,
		) => void | Promise<void>;
	};

	let {
		status,
		readonly = false,
		connecting = false,
		disconnecting = false,
		archiving = false,
		configuring = false,
		serverError = null,
		onConnect,
		onDisconnect,
		onArchive,
		onConfigure,
	}: Props = $props();

	const busy = $derived(connecting || disconnecting || archiving || configuring);
	const connectedAtLabel = $derived(formatConnectedAt(status.connectedAt));
</script>

<div class="overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/35">
	<div class="border-b border-zinc-800 p-5 sm:p-6">
		<div class="flex flex-wrap items-start justify-between gap-3">
			<span class="flex size-10 items-center justify-center rounded-xl bg-sky-500/10 text-sky-300">
				{#if status.connected}
					<Cloud class="size-5" />
				{:else}
					<CloudOff class="size-5" />
				{/if}
			</span>
			<span
				class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium {driveConnectionBadgeClass(
					status,
				)}"
			>
				<span class="size-1.5 rounded-full bg-current"></span>
				{driveConnectionLabel(status)}
			</span>
		</div>
		<div class="mt-4">
			<h2 class="text-sm font-semibold text-zinc-100">Google Drive archive</h2>
			<p class="mt-1 text-xs leading-5 text-zinc-500">
				{#if status.connectionError}
					The saved connection needs attention before archiving can continue.
				{:else if !status.configured}
					Add Google OAuth credentials, then connect an account for off-site storage.
				{:else if status.connected}
					New recordings are archived every five minutes.
				{:else}
					Connect an account to enable automatic off-site archives.
				{/if}
			</p>
		</div>
	</div>

	<div class="p-5 sm:p-6">
		<dl class="grid gap-4 sm:grid-cols-2 xl:grid-cols-1 2xl:grid-cols-2">
			<div class="min-w-0">
				<dt class="text-[11px] font-medium tracking-wide text-zinc-500 uppercase">Account</dt>
				<dd class="mt-1 truncate text-sm text-zinc-200" title={status.accountEmail ?? undefined}>
					{#if status.connected && status.accountEmail}
						{status.accountEmail}
					{:else if status.connected}
						<span class="text-zinc-400">Connected</span>
					{:else}
						<span class="text-zinc-500">—</span>
					{/if}
				</dd>
			</div>

			<div class="min-w-0">
				<dt class="text-[11px] font-medium tracking-wide text-zinc-500 uppercase">Connected</dt>
				<dd class="mt-1 text-sm text-zinc-200">
					{#if connectedAtLabel}
						{connectedAtLabel}
					{:else}
						<span class="text-zinc-500">—</span>
					{/if}
				</dd>
			</div>

			{#if status.folderId}
				<div class="min-w-0 sm:col-span-2 xl:col-span-1 2xl:col-span-2">
					<dt class="text-[11px] font-medium tracking-wide text-zinc-500 uppercase">
						Archive folder
					</dt>
					<dd class="mt-1 truncate font-mono text-xs text-zinc-400" title={status.folderId}>
						{status.folderId}
					</dd>
				</div>
			{/if}
		</dl>

		{#if serverError}
			<p
				class="mt-5 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
				role="alert"
			>
				{serverError}
			</p>
		{/if}
		{#if status.connectionError}
			<p
				class="mt-5 rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-200"
				role="status"
			>
				{status.connectionError}
			</p>
		{/if}

		<div class="mt-5 flex flex-wrap items-center gap-2">
			{#if readonly}
				<p class="text-xs leading-5 text-zinc-500">
					Only administrators can manage Drive archives.
				</p>
			{:else if status.connectionError}
				{#if status.configured}
					<button
						type="button"
						disabled={busy}
						onclick={() => void onConnect()}
						class="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-500 px-3.5 py-2.5 text-sm font-medium text-zinc-950 transition-colors hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60"
					>
						<Link2 class="size-4" />
						{connecting ? "Redirecting…" : "Reconnect Google Drive"}
					</button>
				{/if}
				<button
					type="button"
					disabled={busy}
					onclick={() => void onDisconnect()}
					class="inline-flex items-center justify-center gap-2 rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2.5 text-sm font-medium text-zinc-200 transition-colors hover:border-zinc-600 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
				>
					<Unplug class="size-4" />
					{disconnecting ? "Clearing…" : "Clear saved connection"}
				</button>
			{:else if !status.configured}
				<p class="text-xs leading-5 text-zinc-500">
					NVR_SECRETS_KEY must be set on the server to encrypt these credentials.
				</p>
			{:else if status.connected}
				<button
					type="button"
					disabled={busy}
					onclick={() => void onArchive()}
					class="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-500 px-3.5 py-2.5 text-sm font-medium text-zinc-950 transition-colors hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60"
				>
					<Archive class="size-4" />
					{archiving ? "Archiving…" : "Run archive now"}
				</button>
				<button
					type="button"
					disabled={busy}
					onclick={() => void onDisconnect()}
					class="inline-flex items-center justify-center gap-2 rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2.5 text-sm font-medium text-zinc-200 transition-colors hover:border-zinc-600 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
				>
					<Unplug class="size-4" />
					{disconnecting ? "Disconnecting…" : "Disconnect"}
				</button>
			{:else}
				<button
					type="button"
					disabled={busy}
					onclick={() => void onConnect()}
					class="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-500 px-3.5 py-2.5 text-sm font-medium text-zinc-950 transition-colors hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60"
				>
					<Link2 class="size-4" />
					{connecting ? "Redirecting…" : "Connect Google Drive"}
				</button>
			{/if}
		</div>

		{#if !readonly && status.configured}
			<details class="group mt-5 border-t border-zinc-800 pt-5">
				<summary
					class="flex cursor-pointer list-none items-center justify-between gap-3 text-xs font-medium text-zinc-400 hover:text-zinc-200"
				>
					Google OAuth configuration
					<ChevronDown class="size-4 transition-transform group-open:rotate-180" />
				</summary>
				<p class="mt-3 text-xs leading-5 text-zinc-500">
					New credentials disconnect the current account and start a new connection flow.
				</p>
				<GoogleDriveConfigurationForm submitting={busy} onSubmit={onConfigure} />
			</details>
		{:else if !readonly}
			<div class="mt-5 border-t border-zinc-800 pt-5">
				<GoogleDriveConfigurationForm submitting={busy} onSubmit={onConfigure} />
			</div>
		{/if}
	</div>
</div>
