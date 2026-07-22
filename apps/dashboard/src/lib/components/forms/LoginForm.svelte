<script lang="ts">
	import {
		fieldErrorsFromZod,
		loginFormSchema,
		type LoginFormValues,
	} from "$lib/auth";

	type Props = {
		submitting?: boolean;
		serverError?: string | null;
		onSubmit: (values: LoginFormValues) => void | Promise<void>;
	};

	let { submitting = false, serverError = null, onSubmit }: Props = $props();

	let username = $state("");
	let password = $state("");
	let fieldErrors = $state<Record<string, string>>({});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = loginFormSchema.safeParse({ username, password });
		if (!parsed.success) {
			fieldErrors = fieldErrorsFromZod(parsed.error);
			return;
		}

		await onSubmit(parsed.data);
	}
</script>

<form class="flex flex-col gap-4" onsubmit={handleSubmit} novalidate>
	<div class="flex flex-col gap-1.5">
		<label for="login-username" class="text-sm font-medium text-zinc-300">Username</label>
		<input
			id="login-username"
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
		<label for="login-password" class="text-sm font-medium text-zinc-300">Password</label>
		<input
			id="login-password"
			name="password"
			type="password"
			autocomplete="current-password"
			bind:value={password}
			disabled={submitting}
			class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
			placeholder="••••••••"
		/>
		{#if fieldErrors.password}
			<p class="text-xs text-red-400">{fieldErrors.password}</p>
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
		{submitting ? "Signing in…" : "Sign in"}
	</button>
</form>
