<script lang="ts">
	import { useQueryClient } from "@tanstack/svelte-query";
	import { Server } from "lucide-svelte";
	import { activeOrigin, changeServer, connection } from "$lib/connection";

	const queryClient = useQueryClient();

	const origin = $derived(activeOrigin());

	function handleChange() {
		changeServer();
		queryClient.clear();
	}
</script>

{#if connection.mode === "remote"}
	<div
		class="mb-4 flex items-center justify-between gap-3 rounded-lg border border-zinc-700/70 bg-zinc-800/40 px-3 py-2"
	>
		<div class="flex min-w-0 items-center gap-2 text-sm text-zinc-300">
			<Server class="size-4 shrink-0 text-emerald-400" />
			<span class="truncate">
				<span class="sr-only">Connected to</span>
				{origin}
			</span>
		</div>
		<button
			type="button"
			class="shrink-0 rounded-md px-2 py-1 text-xs font-medium text-zinc-300 transition-colors hover:bg-zinc-700 hover:text-zinc-100"
			onclick={handleChange}
		>
			Change server
		</button>
	</div>
{/if}