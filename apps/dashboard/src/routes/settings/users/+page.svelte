<script lang="ts">
	import { goto } from "$app/navigation";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { authKeys, loadStatus } from "$lib/auth";
	import Spinner from "$lib/components/Spinner.svelte";
	import CreateUserForm from "$lib/components/users/CreateUserForm.svelte";
	import UsersTable from "$lib/components/users/UsersTable.svelte";
	import {
		type CreateUserFormValues,
		createUser,
		deleteUser,
		listUsers,
		type UserPublic,
		UserApiError,
		userKeys,
	} from "$lib/users";

	const queryClient = useQueryClient();

	const statusQuery = createQuery(() => ({
		queryKey: authKeys.status,
		queryFn: loadStatus,
	}));

	const isAdmin = $derived(statusQuery.data?.user?.role === "admin");
	const currentUserId = $derived(statusQuery.data?.user?.id);

	// Non-admin redirect once auth status is known
	$effect(() => {
		if (statusQuery.isSuccess && statusQuery.data?.user && !isAdmin) {
			void goto("/settings", { replaceState: true });
		}
	});

	const usersQuery = createQuery(() => ({
		queryKey: userKeys.list(),
		queryFn: listUsers,
		enabled: isAdmin,
	}));

	let createError = $state<string | null>(null);
	let listError = $state<string | null>(null);
	let deletingId = $state<string | null>(null);

	const createUserMutation = createMutation(() => ({
		mutationFn: (values: CreateUserFormValues) => createUser(values),
		onSuccess: async () => {
			createError = null;
			await queryClient.invalidateQueries({ queryKey: userKeys.all });
		},
		onError: (error: unknown) => {
			if (error instanceof UserApiError) {
				createError = error.message;
				return;
			}
			createError = error instanceof Error ? error.message : "Failed to create user";
		},
	}));

	const deleteUserMutation = createMutation(() => ({
		mutationFn: (id: string) => deleteUser(id),
		onMutate: (id: string) => {
			deletingId = id;
			listError = null;
		},
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: userKeys.all });
		},
		onError: (error: unknown) => {
			if (error instanceof UserApiError) {
				listError = error.message;
				return;
			}
			listError = error instanceof Error ? error.message : "Failed to delete user";
		},
		onSettled: () => {
			deletingId = null;
		},
	}));

	const users = $derived(usersQuery.data ?? []);

	async function handleCreate(values: CreateUserFormValues) {
		createError = null;
		await createUserMutation.mutateAsync(values);
	}

	async function handleDelete(user: UserPublic) {
		const confirmed = window.confirm(
			`Delete user “${user.username}”? This cannot be undone.`,
		);
		if (!confirmed) {
			return;
		}
		await deleteUserMutation.mutateAsync(user.id);
	}
</script>

<svelte:head>
	<title>Users · Settings · NVR</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-3xl flex-col gap-6">
	{#if statusQuery.isPending || (statusQuery.isSuccess && !isAdmin)}
		<div class="flex min-h-[240px] items-center justify-center">
			<Spinner label={isAdmin ? "Loading" : "Redirecting"} />
		</div>
	{:else if isAdmin}
		<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5 sm:p-6">
			<div class="mb-4 flex flex-col gap-1">
				<h2 class="text-sm font-semibold text-zinc-100">Create user</h2>
				<p class="text-xs text-zinc-500">
					Admins can manage settings; operators and viewers have limited access.
				</p>
			</div>
			<CreateUserForm
				submitting={createUserMutation.isPending}
				serverError={createError}
				onSubmit={handleCreate}
			/>
		</div>

		<div class="flex flex-col gap-3">
			<div class="flex items-center justify-between gap-3">
				<h2 class="text-sm font-semibold text-zinc-100">
					All users
					{#if usersQuery.isSuccess}
						<span class="font-normal text-zinc-500">({users.length})</span>
					{/if}
				</h2>
			</div>

			{#if listError}
				<p
					class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
					role="alert"
				>
					{listError}
				</p>
			{/if}

			{#if usersQuery.isPending}
				<div class="flex min-h-[160px] items-center justify-center">
					<Spinner label="Loading users" />
				</div>
			{:else if usersQuery.isError}
				<div
					class="flex flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-10 text-center"
				>
					<p class="text-sm font-medium text-red-200">Could not load users</p>
					<p class="max-w-sm text-sm text-red-300/80">
						{usersQuery.error instanceof Error
							? usersQuery.error.message
							: "Unknown error while loading users."}
					</p>
					<button
						type="button"
						class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
						onclick={() => usersQuery.refetch()}
					>
						Retry
					</button>
				</div>
			{:else if users.length === 0}
				<div
					class="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/40 px-6 py-12 text-center text-sm text-zinc-500"
				>
					No users found.
				</div>
			{:else}
				<UsersTable
					{users}
					{currentUserId}
					{deletingId}
					onDelete={handleDelete}
				/>
			{/if}
		</div>
	{/if}
</section>
