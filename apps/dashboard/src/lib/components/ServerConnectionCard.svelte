<script lang="ts">
	import { connection, connectServer } from "$lib/connection";
	import AuthCard from "./AuthCard.svelte";

	type Props = {
		onConnected: () => void | Promise<void>;
	};

	let { onConnected }: Props = $props();

	let field = $state(connection.prefill);

	async function handleSubmit(event: Event) {
		event.preventDefault();
		if (connection.connecting) {
			return;
		}
		const connected = await connectServer(field);
		if (connected) {
			await onConnected();
		}
	}
</script>

<AuthCard
	title="Connect to a recorder"
	description="Enter the address of your NVR recorder to continue. The dashboard will verify it before connecting."
>
	<form class="flex flex-col gap-4" onsubmit={handleSubmit} novalidate>
		<div class="flex flex-col gap-1.5">
			<label for="server-url" class="text-sm font-medium text-zinc-300">Server URL</label>
			<input
				id="server-url"
				name="serverUrl"
				type="url"
				inputmode="url"
				autocomplete="url"
				bind:value={field}
				disabled={connection.connecting}
				class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
				placeholder="https://recorder.example.com"
			/>
			<p class="text-xs text-zinc-500">
				Hosted tunnels work too — use the HTTPS tunnel address, e.g.
				<code class="rounded bg-zinc-800 px-1 py-0.5 text-zinc-300">https://abc.tunnel.example</code>.
			</p>
		</div>

		{#if connection.error}
			<p
				class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
				role="alert"
			>
				{connection.error}
			</p>
		{/if}

		<button
			type="submit"
			disabled={connection.connecting}
			class="mt-1 inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
		>
			{connection.connecting ? "Connecting…" : "Connect"}
		</button>
	</form>
</AuthCard>