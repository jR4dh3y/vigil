<script lang="ts">
	import {
		isSubRouteActive,
		type MainSection,
		type SubRoute,
	} from "$lib/nav/subroutes";

	type Props = {
		section: MainSection;
		pathname: string;
		isAdmin?: boolean;
	};

	let { section, pathname, isAdmin = false }: Props = $props();

	const visibleRoutes = $derived(
		section.subRoutes.filter((route: SubRoute) => !route.adminOnly || isAdmin),
	);
</script>

<header
	class="flex h-14 shrink-0 items-center gap-4 border-b border-zinc-800/80 bg-zinc-950/90 px-4 backdrop-blur-md sm:px-6"
>
	<h1 class="text-sm font-semibold tracking-tight text-zinc-100">{section.title}</h1>

	{#if visibleRoutes.length > 0}
		<nav class="flex flex-1 items-center gap-1" aria-label="Section">
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
</header>
