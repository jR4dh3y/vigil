<script lang="ts">
	import type { Snippet } from "svelte";
	import type { UserPublic } from "$lib/auth";
	import { resolveMainSection } from "$lib/nav/subroutes";
	import Sidebar from "./Sidebar.svelte";
	import SubNav from "./SubNav.svelte";

	type Props = {
		user: UserPublic;
		pathname: string;
		loggingOut?: boolean;
		onLogout: () => void;
		children: Snippet;
	};

	let { user, pathname, loggingOut = false, onLogout, children }: Props = $props();

	const section = $derived(resolveMainSection(pathname));
	const isAdmin = $derived(user.role === "admin");
	const isLive = $derived(pathname === "/");
</script>

<div class="flex h-screen overflow-hidden bg-zinc-950 text-zinc-100">
	<Sidebar {user} {pathname} {loggingOut} {onLogout} />

	<div class="flex min-w-0 flex-1 flex-col">
		{#if section}
			<SubNav {section} {pathname} {isAdmin} />
		{/if}

		<main
			class="min-h-0 flex-1 overflow-auto
				{isLive ? 'p-0' : 'px-4 py-6 sm:px-6'}"
		>
			{#if isLive}
				{@render children()}
			{:else}
				<div class="mx-auto w-full max-w-7xl">
					{@render children()}
				</div>
			{/if}
		</main>
	</div>
</div>
