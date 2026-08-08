<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/state";
	import { useQueryClient } from "@tanstack/svelte-query";
	import type { Snippet } from "svelte";
	import { connection, initConnection } from "$lib/connection";
	import ServerConnectionCard from "./ServerConnectionCard.svelte";
	import Spinner from "./Spinner.svelte";

	type Props = {
		children: Snippet;
	};

	let { children }: Props = $props();

	const queryClient = useQueryClient();

	let initialized = $state(false);

	onMount(() => {
		const prefill = page.url.searchParams.get("server");
		void initConnection(prefill).then(() => {
			initialized = true;
		});
	});

	async function handleConnected() {
		// A server (re)connect invalidates any cached data from a previous one.
		queryClient.clear();
	}
</script>

{#if !initialized || connection.mode === "detecting"}
	<div class="flex min-h-screen items-center justify-center bg-zinc-950">
		<Spinner label="Looking for a recorder" />
	</div>
{:else if connection.mode === "none"}
	<ServerConnectionCard onConnected={handleConnected} />
{:else}
	{@render children()}
{/if}