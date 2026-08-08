<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import type { Snippet } from "svelte";
	import {
		authKeys,
		isPublicRoute,
		loadStatus,
		logout as logoutRequest,
	} from "$lib/auth";
	import { resolveAuthRedirect } from "$lib/auth/guard";
	import { changeServer, connection } from "$lib/connection";
	import AppShell from "./AppShell.svelte";
	import Spinner from "./Spinner.svelte";

	type Props = {
		children: Snippet;
	};

	let { children }: Props = $props();

	const queryClient = useQueryClient();

	const statusQuery = createQuery(() => ({
		queryKey: authKeys.status,
		queryFn: loadStatus,
		retry: false,
	}));

	const logoutMutation = createMutation(() => ({
		mutationFn: logoutRequest,
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: authKeys.all });
			await goto(resolve("/login"));
		},
	}));

	const pathname = $derived(page.url.pathname);
	const status = $derived(statusQuery.data);
	const isLoading = $derived(statusQuery.isPending);
	const isError = $derived(statusQuery.isError);
	const user = $derived(status?.user);
	const showShell = $derived(Boolean(user) && !isPublicRoute(pathname));

	$effect(() => {
		if (isLoading || isError || !status) {
			return;
		}
		const redirect = resolveAuthRedirect(status, pathname);
		if (redirect.kind === "goto") {
			void goto(resolve(redirect.to), { replaceState: true });
		}
	});

	function handleLogout() {
		logoutMutation.mutate();
	}

	function handleChangeServer() {
		changeServer();
		queryClient.clear();
	}
</script>

{#if isLoading}
	<div class="flex min-h-screen items-center justify-center bg-zinc-950">
		<Spinner label="Checking session" />
	</div>
{:else if isError}
	<div
		class="flex min-h-screen flex-col items-center justify-center gap-4 bg-zinc-950 px-4 text-center text-zinc-100"
	>
		<p class="text-lg font-medium text-zinc-200">Could not reach the API</p>
		<p class="max-w-sm text-sm text-zinc-500">
			{statusQuery.error instanceof Error
				? statusQuery.error.message
				: "Unknown error while loading auth status."}
		</p>
		<div class="flex gap-3">
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => statusQuery.refetch()}
			>
				Retry
			</button>
			{#if connection.mode === "remote"}
				<button
					type="button"
					class="rounded-lg border border-zinc-700 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-800"
					onclick={handleChangeServer}
				>
					Change server
				</button>
			{/if}
		</div>
	</div>
{:else if showShell && user}
	<AppShell
		{user}
		{pathname}
		loggingOut={logoutMutation.isPending}
		onLogout={handleLogout}
	>
		{@render children()}
	</AppShell>
{:else}
	{@render children()}
{/if}
