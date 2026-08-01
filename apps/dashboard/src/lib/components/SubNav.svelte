<script lang="ts">
	import {
		isSubRouteActive,
		type MainSection,
		type SubRoute,
	} from "$lib/nav/subroutes";
	import {
		PAGE_ACTIONS_HOST_ID,
		PAGE_META_HOST_ID,
	} from "$lib/nav/page-actions.svelte";
	import { isCameraDetailRoute } from "$lib/nav/camera-route";

	type Props = {
		section: MainSection;
		pathname: string;
		isAdmin?: boolean;
	};

	let { section, pathname, isAdmin = false }: Props = $props();

	const visibleRoutes = $derived(
		section.subRoutes.filter(
			(route: SubRoute) =>
				!route.adminOnly || isAdmin,
		),
	);
	const isCameraDetail = $derived(isCameraDetailRoute(pathname));
</script>

<header
	class="flex h-14 shrink-0 items-center gap-4 border-b border-zinc-800/80 bg-zinc-950/90 px-4 backdrop-blur-md sm:px-6"
>
	{#if !isCameraDetail}
		<h1 class="shrink-0 text-sm font-semibold tracking-tight text-zinc-100">{section.title}</h1>
	{/if}

	<!-- Left-side meta (e.g. focused camera label) -->
	<div
		id={PAGE_META_HOST_ID}
		class="flex min-w-0 items-center gap-2 empty:hidden"
	></div>

	{#if visibleRoutes.length > 0 && !isCameraDetail}
		<nav class="flex items-center gap-1" aria-label="Section">
			{#each visibleRoutes as route (route.href)}
				{@const active = isSubRouteActive(pathname, route)}
				<a
					href={route.href}
					class="rounded-md px-2.5 py-1.5 text-sm transition-colors no-underline
						{active
						? 'bg-zinc-800 text-zinc-100'
						: 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'}"
					aria-current={active ? "page" : undefined}
				>
					{route.label}
				</a>
			{/each}
		</nav>
	{/if}

	<!-- Right-side actions -->
	<div
		id={PAGE_ACTIONS_HOST_ID}
		class="ml-auto flex min-w-0 flex-wrap items-center justify-end gap-2 empty:hidden"
	></div>
</header>
