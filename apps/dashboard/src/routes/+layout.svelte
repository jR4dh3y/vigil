<script lang="ts">
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import AuthGate from "$lib/components/AuthGate.svelte";
	import ServerGate from "$lib/components/ServerGate.svelte";
	import { createQueryClient } from "$lib/query";
	import "../app.css";

	let { children } = $props();

	const queryClient = createQueryClient();
</script>

<svelte:head>
	<link rel="icon" type="image/png" href="/owl.png" />
	<title>Vigil</title>
</svelte:head>

<QueryClientProvider client={queryClient}>
	<ServerGate>
		<AuthGate>
			{@render children()}
		</AuthGate>
	</ServerGate>
</QueryClientProvider>
