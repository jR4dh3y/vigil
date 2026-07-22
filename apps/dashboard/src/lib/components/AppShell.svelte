<script lang="ts">
	import type { Snippet } from "svelte";
	import type { UserPublic } from "$lib/auth";
	import Nav from "./Nav.svelte";

	type Props = {
		user: UserPublic;
		pathname: string;
		loggingOut?: boolean;
		onLogout: () => void;
		children: Snippet;
	};

	let { user, pathname, loggingOut = false, onLogout, children }: Props = $props();
</script>

<div class="flex min-h-screen flex-col bg-zinc-950 text-zinc-100">
	<Nav {user} {pathname} {loggingOut} {onLogout} />
	<main class="mx-auto w-full max-w-7xl flex-1 px-4 py-6 sm:px-6">
		{@render children()}
	</main>
</div>
