<script lang="ts">
	import { KeyRound } from "lucide-svelte";
	import { z } from "zod";
	import type { GDriveConfigurationRequest } from "$lib/storage";

	type Props = {
		submitting?: boolean;
		onSubmit: (values: GDriveConfigurationRequest) => void | Promise<void>;
	};

	let { submitting = false, onSubmit }: Props = $props();

	const configurationSchema = z.object({
		clientId: z.string().trim().min(1, "Google OAuth client ID is required"),
		clientSecret: z.string().trim().min(1, "Google OAuth client secret is required"),
		redirectUrl: z
			.url("Enter an absolute HTTP(S) redirect URL")
			.refine(
				(value) => new URL(value).protocol === "http:" || new URL(value).protocol === "https:",
				{
					message: "Enter an absolute HTTP(S) redirect URL",
				},
			),
	});

	let clientId = $state("");
	let clientSecret = $state("");
	let redirectUrl = $state("");
	let fieldErrors = $state<Record<string, string>>({});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};

		const parsed = configurationSchema.safeParse({
			clientId,
			clientSecret,
			redirectUrl,
		});
		if (!parsed.success) {
			for (const issue of parsed.error.issues) {
				const field = issue.path[0];
				if (typeof field === "string" && fieldErrors[field] === undefined) {
					fieldErrors[field] = issue.message;
				}
			}
			return;
		}

		await onSubmit(parsed.data);
	}
</script>

<form class="mt-5 flex flex-col gap-4" onsubmit={handleSubmit} novalidate>
	<div class="flex flex-col gap-1.5">
		<label for="gdrive-client-id" class="text-xs font-medium text-zinc-400">Client ID</label>
		<input
			id="gdrive-client-id"
			name="clientId"
			type="text"
			bind:value={clientId}
			disabled={submitting}
			autocomplete="off"
			spellcheck="false"
			class="h-10 rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 font-mono text-sm text-zinc-100 outline-none ring-emerald-500/30 transition placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
			placeholder="1234567890-abc.apps.googleusercontent.com"
		/>
		{#if fieldErrors.clientId}
			<p class="text-xs text-red-400">{fieldErrors.clientId}</p>
		{/if}
	</div>

	<div class="flex flex-col gap-1.5">
		<label for="gdrive-client-secret" class="text-xs font-medium text-zinc-400">Client secret</label
		>
		<input
			id="gdrive-client-secret"
			name="clientSecret"
			type="password"
			bind:value={clientSecret}
			disabled={submitting}
			autocomplete="new-password"
			class="h-10 rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 font-mono text-sm text-zinc-100 outline-none ring-emerald-500/30 transition placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
		/>
		<p class="text-xs text-zinc-500">Encrypted at rest and never shown again.</p>
		{#if fieldErrors.clientSecret}
			<p class="text-xs text-red-400">{fieldErrors.clientSecret}</p>
		{/if}
	</div>

	<div class="flex flex-col gap-1.5">
		<label for="gdrive-redirect-url" class="text-xs font-medium text-zinc-400">Redirect URL</label>
		<input
			id="gdrive-redirect-url"
			name="redirectUrl"
			type="url"
			bind:value={redirectUrl}
			disabled={submitting}
			autocomplete="url"
			spellcheck="false"
			class="h-10 rounded-lg border border-zinc-700 bg-zinc-950/80 px-3 font-mono text-sm text-zinc-100 outline-none ring-emerald-500/30 transition placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2 disabled:cursor-not-allowed disabled:opacity-60"
			placeholder="https://nvr.example.com/api/v1/storage/gdrive/callback"
		/>
		<p class="text-xs text-zinc-500">
			This must exactly match an authorized redirect URI in Google Cloud Console.
		</p>
		{#if fieldErrors.redirectUrl}
			<p class="text-xs text-red-400">{fieldErrors.redirectUrl}</p>
		{/if}
	</div>

	<button
		type="submit"
		disabled={submitting}
		class="inline-flex w-fit items-center justify-center gap-2 rounded-lg bg-emerald-500 px-3.5 py-2.5 text-sm font-medium text-zinc-950 transition-colors hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60"
	>
		<KeyRound class="size-4" />
		{submitting ? "Saving…" : "Save and connect Google Drive"}
	</button>
</form>
