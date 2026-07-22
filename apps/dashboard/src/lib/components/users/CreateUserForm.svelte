<script lang="ts">
	import {
		type CreateUserFormValues,
		createUserFormSchema,
		fieldErrorsFromZod,
		userRoles,
	} from "$lib/users";

	type Props = {
		submitting?: boolean;
		serverError?: string | null;
		onSubmit: (values: CreateUserFormValues) => void | Promise<void>;
	};

	let { submitting = false, serverError = null, onSubmit }: Props = $props();

	let username = $state("");
	let password = $state("");
	let role = $state<CreateUserFormValues["role"]>("viewer");
	let fieldErrors = $state<Record<string, string>>({});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = createUserFormSchema.safeParse({ username, password, role });
		if (!parsed.success) {
			fieldErrors = fieldErrorsFromZod(parsed.error);
			return;
		}

		await onSubmit(parsed.data);
		username = "";
		password = "";
		role = "viewer";
	}
</script>

<form class="flex flex-col gap-4" onsubmit={handleSubmit} novalidate>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex flex-col gap-1.5">
			<label for="create-user-username" class="text-sm font-medium text-zinc-300"
				>Username</label
			>
			<input
				id="create-user-username"
				name="username"
				type="text"
				autocomplete="off"
				bind:value={username}
				disabled={submitting}
				class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
				placeholder="operator"
			/>
			{#if fieldErrors.username}
				<p class="text-xs text-red-400">{fieldErrors.username}</p>
			{/if}
		</div>

		<div class="flex flex-col gap-1.5">
			<label for="create-user-role" class="text-sm font-medium text-zinc-300">Role</label>
			<select
				id="create-user-role"
				name="role"
				bind:value={role}
				disabled={submitting}
				class="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 focus:border-emerald-500/60 focus:ring-2 disabled:opacity-50"
			>
				{#each userRoles as r (r)}
					<option value={r}>{r}</option>
				{/each}
			</select>
			{#if fieldErrors.role}
				<p class="text-xs text-red-400">{fieldErrors.role}</p>
			{/if}
		</div>
	</div>

	<div class="flex flex-col gap-1.5">
		<label for="create-user-password" class="text-sm font-medium text-zinc-300"
			>Password</label
		>
		<input
			id="create-user-password"
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
		class="inline-flex w-fit items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
	>
		{submitting ? "Creating…" : "Create user"}
	</button>
</form>
