<script lang="ts">
	import {
		fieldErrorsFromZod,
		setupFormSchema,
		type SetupFormValues,
	} from "$lib/auth";

	type Props = {
		submitting?: boolean;
		serverError?: string | null;
		onSubmit: (values: SetupFormValues) => void | Promise<void>;
	};

	let { submitting = false, serverError = null, onSubmit }: Props = $props();

	let username = $state("");
	let password = $state("");
	let fieldErrors = $state<Record<string, string>>({});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = setupFormSchema.safeParse({ username, password });
		if (!parsed.success) {
			fieldErrors = fieldErrorsFromZod(parsed.error);
			return;
		}

		await onSubmit(parsed.data);
	}
</script>

<form class="flex flex-col gap-4" onsubmit={handleSubmit} novalidate>
	<div class="flex flex-col gap-1.5">
		<label for="setup-username" class="text-sm font-medium text-zinc-300">Admin username</label>
		<input
			id="setup-username"
			name="username"
			type="text"
			autocomplete="username"
			bind:value={username}
			disabled={submitting}
			class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
			placeholder="admin"
		/>
		{#if fieldErrors.username}
			<p class="text-xs text-red-400">{fieldErrors.username}</p>
		{/if}
	</div>

	<div class="flex flex-col gap-1.5">
		<label for="setup-password" class="text-sm font-medium text-zinc-300">Password</label>
		<input
			id="setup-password"
			name="password"
			type="password"
			autocomplete="new-password"
			bind:value={password}
			disabled={submitting}
			class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
			placeholder="At least 8 characters"
		/>
		{#if fieldErrors.password}
			<p class="text-xs text-red-400">{fieldErrors.password}</p>
		{:else}
			<p class="text-xs text-zinc-500">Minimum 8 characters.</p>
		{/if}
	</div>

	{#if serverError}
		<p
			class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{serverError}
		</p>
	{/if}

	<button
		type="submit"
		disabled={submitting}
		class="mt-1 inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
	>
		{submitting ? "Creating account…" : "Create admin account"}
	</button>
</form>
