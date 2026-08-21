<script lang="ts">
	import type { Snippet } from "svelte";

	type Props = {
		title: string;
		description?: string;
		children: Snippet;
	};

	let { title, description, children }: Props = $props();
</script>

<div class="flex min-h-screen items-center justify-center bg-zinc-950 px-4 py-12 text-zinc-100">
	<div class="w-full max-w-md">
		<div class="mb-8 flex flex-col items-center gap-3 text-center">
			<span
				class="flex size-12 items-center justify-center overflow-hidden rounded-xl bg-[#f3eee3] p-0.5"
			>
				<img src="/owl.png" alt="Vigil owl mark" width="1254" height="1254" class="size-full object-contain" />
			</span>
			<div>
				<p class="text-xs font-medium uppercase tracking-[0.2em] text-zinc-500">Vigil</p>
				<h1 class="mt-1 text-2xl font-semibold tracking-tight text-zinc-50">{title}</h1>
				{#if description}
					<p class="mt-2 text-sm text-zinc-400">{description}</p>
				{/if}
			</div>
		</div>

		<div class="rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 shadow-xl shadow-black/40">
			{@render children()}
		</div>
	</div>
</div>
