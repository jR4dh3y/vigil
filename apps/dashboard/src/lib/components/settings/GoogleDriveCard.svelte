<script lang="ts">
	import type { GDriveStatus } from "$lib/storage";
	import {
		driveConnectionBadgeClass,
		driveConnectionLabel,
		formatConnectedAt,
	} from "$lib/storage";

	type Props = {
		status: GDriveStatus;
		readonly?: boolean;
		connecting?: boolean;
		disconnecting?: boolean;
		archiving?: boolean;
		serverError?: string | null;
		onConnect: () => void | Promise<void>;
		onDisconnect: () => void | Promise<void>;
		onArchive: () => void | Promise<void>;
	};

	let {
		status,
		readonly = false,
		connecting = false,
		disconnecting = false,
		archiving = false,
		serverError = null,
		onConnect,
		onDisconnect,
		onArchive,
	}: Props = $props();

	const busy = $derived(connecting || disconnecting || archiving);
	const connectedAtLabel = $derived(formatConnectedAt(status.connectedAt));
</script>

<div class="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 sm:p-6">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div class="flex flex-col gap-1">
			<h2 class="text-sm font-semibold text-zinc-100">Google Drive</h2>
			<p class="text-xs text-zinc-500">
				{#if status.connectionError}
					The saved connection needs attention before archiving can continue.
				{:else if !status.configured}
					Server credentials are not set. Configure NVR_GOOGLE_CLIENT_ID/SECRET/REDIRECT_URL
					and NVR_SECRETS_KEY on the NVR host before connecting.
				{:else if status.connected}
					Unarchived recordings upload to Drive daily at 00:00 UTC, or use Run archive
					now.
				{:else}
					Connect a Google account to enable Drive archive for this NVR.
				{/if}
			</p>
		</div>
		<span
			class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium {driveConnectionBadgeClass(
				status,
			)}"
		>
			{driveConnectionLabel(status)}
		</span>
	</div>

	<dl class="mt-5 grid gap-4 sm:grid-cols-2">
		<div class="flex flex-col gap-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Account</dt>
			<dd class="text-sm text-zinc-200">
				{#if status.connected && status.accountEmail}
					{status.accountEmail}
				{:else if status.connected}
					<span class="text-zinc-400">Connected</span>
				{:else}
					<span class="text-zinc-500">—</span>
				{/if}
			</dd>
		</div>

		<div class="flex flex-col gap-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Connected</dt>
			<dd class="text-sm text-zinc-200">
				{#if connectedAtLabel}
					{connectedAtLabel}
				{:else}
					<span class="text-zinc-500">—</span>
				{/if}
			</dd>
		</div>

		{#if status.folderId}
			<div class="flex flex-col gap-1 sm:col-span-2">
				<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Folder ID</dt>
				<dd class="truncate font-mono text-xs text-zinc-400" title={status.folderId}>
					{status.folderId}
				</dd>
			</div>
		{/if}
	</dl>

		{#if serverError}
		<p
			class="mt-4 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{serverError}
		</p>
	{/if}
	{#if status.connectionError}
		<p
			class="mt-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-200"
			role="status"
		>
			{status.connectionError}
		</p>
	{/if}

	<div class="mt-5 flex flex-wrap items-center gap-2">
		{#if readonly}
			<p class="text-xs text-zinc-500">Only administrators can connect or disconnect Drive.</p>
		{:else if status.connectionError}
			{#if status.configured}
				<button
					type="button"
					disabled={busy}
					onclick={() => void onConnect()}
					class="inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
				>
					{connecting ? "Redirecting…" : "Reconnect Google Drive"}
				</button>
			{/if}
			<button
				type="button"
				disabled={busy}
				onclick={() => void onDisconnect()}
				class="inline-flex items-center justify-center rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2.5 text-sm font-medium text-zinc-200 transition-colors hover:border-zinc-600 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
			>
				{disconnecting ? "Clearing…" : "Clear saved connection"}
			</button>
		{:else if !status.configured}
			<p class="text-xs text-zinc-500">
				Configure Google OAuth on the server, then refresh this page.
			</p>
		{:else if status.connected}
			<button
				type="button"
				disabled={busy}
				onclick={() => void onArchive()}
				class="inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
			>
				{archiving ? "Archiving…" : "Run archive now"}
			</button>
			<button
				type="button"
				disabled={busy}
				onclick={() => void onDisconnect()}
				class="inline-flex items-center justify-center rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2.5 text-sm font-medium text-zinc-200 transition-colors hover:border-zinc-600 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
			>
				{disconnecting ? "Disconnecting…" : "Disconnect"}
			</button>
		{:else}
			<button
				type="button"
				disabled={busy}
				onclick={() => void onConnect()}
				class="inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
			>
				{connecting ? "Redirecting…" : "Connect Google Drive"}
			</button>
		{/if}
	</div>
</div>
