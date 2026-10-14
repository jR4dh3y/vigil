<script lang="ts">
	import { Archive, ChevronDown, Link2, Unplug } from "lucide-svelte";
	import type { GDriveStatus } from "$lib/storage";
	import GoogleDriveConfigurationForm from "./GoogleDriveConfigurationForm.svelte";

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
</script>

<section aria-label="Google Drive archive" class="min-w-0">
	<div class="mb-5 flex flex-wrap items-baseline justify-between gap-3">
		<h2 class="text-sm font-semibold text-zinc-100">Google Drive archive</h2>
	</div>

	<div>
		{#if status.connected}
			<p class="break-all text-sm text-zinc-200">{status.accountEmail || "Account connected"}</p>
			<p class="mt-1 text-xs text-zinc-400">Archives new recordings every five minutes.</p>
		{:else if !status.configured}
			<p class="text-sm text-zinc-400">Add OAuth credentials to connect an account.</p>
		{:else if !status.connectionError}
			<p class="text-sm text-zinc-400">Connect an account to archive recordings.</p>
		{/if}
		{#if status.folderId}
			<details class="mt-4">
				<summary class="cursor-pointer text-xs text-zinc-400 hover:text-zinc-200">Archive folder</summary>
				<p class="mt-2 break-all font-mono text-xs text-zinc-400">{status.folderId}</p>
			</details>
		{/if}

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
				<p class="text-xs leading-5 text-zinc-400">
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
				<p class="text-xs leading-5 text-zinc-400">
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
				<p class="mt-3 text-xs leading-5 text-zinc-400">
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
</section>
