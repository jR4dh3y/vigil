<script lang="ts">
	import { Camera, LogOut, Settings, Video, Zap } from "lucide-svelte";
	import type { UserPublic } from "$lib/auth";

	type NavItem = {
		href: string;
		label: string;
		icon: typeof Video;
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

<aside
	class="flex w-14 shrink-0 flex-col border-r border-zinc-800/80 bg-zinc-950 sm:w-52"
	aria-label="Main navigation"
>
	<a
		href="/"
		class="flex h-14 items-center gap-2.5 border-b border-zinc-800/80 px-3 text-zinc-100 no-underline sm:px-4"
	>
		<span
			class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-emerald-500/15 text-emerald-400"
		>
			<Camera class="size-4" />
		</span>
		<span class="hidden text-sm font-semibold tracking-wide sm:inline">NVR</span>
	</a>

	<nav class="flex flex-1 flex-col gap-0.5 p-2" aria-label="Main">
		{#each items as item (item.href)}
			<a
				href={item.href}
				class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm transition-colors no-underline
					{isActive(item.href)
					? 'bg-zinc-800 text-zinc-100'
					: 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}"
				aria-current={isActive(item.href) ? "page" : undefined}
				title={item.label}
			>
				<item.icon class="size-4 shrink-0" />
				<span class="hidden truncate sm:inline">{item.label}</span>
			</a>
		{/each}
	</nav>

	<div class="flex flex-col gap-2 border-t border-zinc-800/80 p-2">
		<div class="hidden px-2.5 py-1 sm:block">
			<p class="truncate text-sm text-zinc-300">{user.username}</p>
			<p class="truncate text-xs capitalize text-zinc-500">{user.role}</p>
		</div>
		<button
			type="button"
			class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-zinc-400 transition-colors hover:bg-zinc-900 hover:text-zinc-200 disabled:cursor-not-allowed disabled:opacity-50"
			disabled={loggingOut}
			onclick={onLogout}
			title={loggingOut ? "Signing out…" : "Log out"}
		>
			<LogOut class="size-4 shrink-0" />
			<span class="hidden sm:inline">{loggingOut ? "Signing out…" : "Log out"}</span>
		</button>
	</div>
</aside>
