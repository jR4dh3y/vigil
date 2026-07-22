<script lang="ts">
	import { Camera, LogOut, Settings, Video, Zap } from "lucide-svelte";
	import type { UserPublic } from "$lib/auth";

	type NavItem = {
		href: string;
		label: string;
		icon: typeof Video;
		disabled?: boolean;
	};

	type Props = {
		user: UserPublic;
		pathname: string;
		loggingOut?: boolean;
		onLogout: () => void;
	};

	let { user, pathname, loggingOut = false, onLogout }: Props = $props();

	const items: NavItem[] = [
		{ href: "/", label: "Live", icon: Video },
		{ href: "/cameras", label: "Cameras", icon: Camera },
		{ href: "/events", label: "Events", icon: Zap },
		{ href: "/settings", label: "Settings", icon: Settings },
	];

	function isActive(href: string): boolean {
		if (href === "/") {
			return pathname === "/";
		}
		return pathname === href || pathname.startsWith(`${href}/`);
	}
</script>

<header
	class="sticky top-0 z-40 border-b border-zinc-800/80 bg-zinc-950/90 backdrop-blur-md"
>
	<div class="mx-auto flex h-14 max-w-7xl items-center gap-6 px-4 sm:px-6">
		<a href="/" class="flex items-center gap-2 text-zinc-100 no-underline">
			<span
				class="flex size-8 items-center justify-center rounded-lg bg-emerald-500/15 text-emerald-400"
			>
				<Camera class="size-4" />
			</span>
			<span class="text-sm font-semibold tracking-wide">NVR</span>
		</a>

		<nav class="flex flex-1 items-center gap-1" aria-label="Main">
			{#each items as item (item.href)}
				{#if item.disabled}
					<span
						class="inline-flex cursor-not-allowed items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-zinc-600"
						title="Coming soon"
					>
						<item.icon class="size-3.5" />
						{item.label}
					</span>
				{:else}
					<a
						href={item.href}
						class="inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm transition-colors no-underline
							{isActive(item.href)
							? 'bg-zinc-800 text-zinc-100'
							: 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}"
						aria-current={isActive(item.href) ? "page" : undefined}
					>
						<item.icon class="size-3.5" />
						{item.label}
					</a>
				{/if}
			{/each}
		</nav>

		<div class="flex items-center gap-3">
			<span class="hidden text-sm text-zinc-500 sm:inline">
				<span class="text-zinc-300">{user.username}</span>
				<span class="mx-1.5 text-zinc-700">·</span>
				<span class="capitalize">{user.role}</span>
			</span>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-50"
				disabled={loggingOut}
				onclick={onLogout}
			>
				<LogOut class="size-3.5" />
				{loggingOut ? "Signing out…" : "Log out"}
			</button>
		</div>
	</div>
</header>
