<script lang="ts">
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import favicon from "$lib/assets/favicon.svg";
	import AuthGate from "$lib/components/AuthGate.svelte";
	import { createQueryClient } from "$lib/query";
	import "../app.css";

	let { children } = $props();

	const queryClient = createQueryClient();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>NVR</title>
</svelte:head>

<QueryClientProvider client={queryClient}>
	<AuthGate>
		{@render children()}
	</AuthGate>
</QueryClientProvider>
