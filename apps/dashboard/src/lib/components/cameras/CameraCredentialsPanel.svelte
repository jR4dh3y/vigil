<script lang="ts">
	import { KeyRound, Radar } from "lucide-svelte";
	import type { DiscoveredCamera } from "$lib/cameras";

	type Props = {
		camera: DiscoveredCamera;
		detecting?: boolean;
		error?: string | null;
		initialUsername?: string;
		initialPassword?: string;
		onDetect: (username: string, password: string) => void | Promise<void>;
		onContinue: (username: string, password: string) => void;
	};

	let {
		camera,
		detecting = false,
		error = null,
		initialUsername = "",
		initialPassword = "",
		onDetect,
		onContinue,
	}: Props = $props();

	let username = $state("");
	let password = $state("");
	let validationError = $state<string | null>(null);

	$effect.pre(() => {
		username = initialUsername;
		password = initialPassword;
	});

	const inputClass =
		"rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-emerald-500/40 placeholder:text-zinc-600 focus:border-emerald-500/60 focus:ring-2";

	function validate(): boolean {
		validationError = null;
		if (!username.trim() || !password) {
			validationError = "Enter the camera username and password to detect its RTSP streams.";
			return false;
		}
		return true;
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();
		if (validate()) {
			await onDetect(username.trim(), password);
		}
	}

	function handleContinue() {
		if (validate()) {
			onContinue(username.trim(), password);
		}
	}
</script>

<form class="flex flex-col gap-5" onsubmit={handleSubmit} novalidate>
	<div>
		<div class="mb-2 flex items-center gap-2 text-emerald-400">
			<KeyRound class="size-4" />
			<span class="text-xs font-semibold uppercase tracking-[0.18em]">Camera credentials</span>
		</div>
		<h1 class="text-xl font-semibold text-zinc-100">Connect to {camera.name}</h1>
		<p class="mt-1 font-mono text-xs text-zinc-500">{camera.host}</p>
		<p class="mt-3 text-sm text-zinc-400">
			The camera needs its username and password before the NVR can ask for authenticated RTSP stream
			URLs.
		</p>
	</div>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex flex-col gap-1.5">
			<label for="discovery-username" class="text-sm font-medium text-zinc-300">Username</label>
			<input
				id="discovery-username"
				type="text"
				autocomplete="username"
				bind:value={username}
				disabled={detecting}
				class={inputClass}
				placeholder="admin"
			/>
		</div>
		<div class="flex flex-col gap-1.5">
			<label for="discovery-password" class="text-sm font-medium text-zinc-300">Password</label>
			<input
				id="discovery-password"
				type="password"
				autocomplete="current-password"
				bind:value={password}
				disabled={detecting}
				class={inputClass}
				placeholder="Camera password"
			/>
		</div>
	</div>

	{#if validationError || error}
		<p
			class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{validationError ?? error}
		</p>
	{/if}

	<div class="flex flex-wrap items-center justify-between gap-3 border-t border-zinc-800 pt-4">
		<button
			type="button"
			disabled={detecting}
			onclick={handleContinue}
			class="rounded-lg border border-zinc-700 bg-zinc-900 px-3.5 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-600 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
		>
			Enter streams manually
		</button>
		<button
			type="submit"
			disabled={detecting}
			class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
		>
			<Radar class="size-4 {detecting ? 'animate-spin' : ''}" />
			{detecting ? "Detecting streams…" : "Detect RTSP streams"}
		</button>
	</div>
</form>
