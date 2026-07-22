<script lang="ts">
	import { Trash2 } from "lucide-svelte";
	import type { UserPublic } from "$lib/users";

	type Props = {
		users: UserPublic[];
		currentUserId?: string;
		deletingId?: string | null;
		onDelete: (user: UserPublic) => void | Promise<void>;
	};

	let { users, currentUserId, deletingId = null, onDelete }: Props = $props();

	function roleBadgeClass(role: UserPublic["role"]): string {
		switch (role) {
			case "admin":
				return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300";
			case "operator":
				return "border-sky-500/30 bg-sky-500/10 text-sky-300";
			default:
				return "border-zinc-600/40 bg-zinc-800/80 text-zinc-400";
		}
	}
</script>

<div class="overflow-hidden rounded-xl border border-zinc-800">
	<table class="w-full border-collapse text-left text-sm">
		<thead class="border-b border-zinc-800 bg-zinc-900/80 text-xs tracking-wide text-zinc-500 uppercase">
			<tr>
				<th class="px-4 py-3 font-medium">Username</th>
				<th class="px-4 py-3 font-medium">Role</th>
				<th class="px-4 py-3 font-medium text-right">Actions</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-zinc-800/80">
			{#each users as user (user.id)}
				<tr class="bg-zinc-900/40 transition-colors hover:bg-zinc-900/70">
					<td class="px-4 py-3">
						<div class="flex items-center gap-2">
							<span class="font-medium text-zinc-100">{user.username}</span>
							{#if user.id === currentUserId}
								<span
									class="rounded border border-zinc-700 bg-zinc-800 px-1.5 py-0.5 text-[10px] text-zinc-400"
								>
									You
								</span>
							{/if}
						</div>
						<p class="mt-0.5 font-mono text-xs text-zinc-600">{user.id}</p>
					</td>
					<td class="px-4 py-3">
						<span
							class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium capitalize {roleBadgeClass(
								user.role,
							)}"
						>
							{user.role}
						</span>
					</td>
					<td class="px-4 py-3 text-right">
						<button
							type="button"
							class="inline-flex items-center gap-1.5 rounded-md border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-xs text-zinc-400 transition-colors hover:border-red-500/40 hover:bg-red-500/10 hover:text-red-300 disabled:cursor-not-allowed disabled:opacity-40"
							disabled={deletingId === user.id || user.id === currentUserId}
							title={user.id === currentUserId
								? "You cannot delete your own account"
								: "Delete user"}
							onclick={() => onDelete(user)}
						>
							<Trash2 class="size-3.5" />
							{deletingId === user.id ? "Deleting…" : "Delete"}
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
