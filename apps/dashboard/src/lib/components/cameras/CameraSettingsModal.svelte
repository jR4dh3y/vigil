<script lang="ts">
	import type { EditCameraFormValues, ProbeCameraRequest, ProbeResult } from "$lib/cameras";
	import CameraForm from "./CameraForm.svelte";

	type Props = {
		open: boolean;
		cameraName: string;
		initial: EditCameraFormValues;
		submitting?: boolean;
		probing?: boolean;
		deleting?: boolean;
		serverError?: string | null;
		probeResult?: ProbeResult | null;
		probeError?: string | null;
		onClose: () => void;
		onSubmit: (values: EditCameraFormValues) => void | Promise<void>;
		onProbe: (input: ProbeCameraRequest) => void | Promise<void>;
		onDelete: () => void | Promise<void>;
	};

	let {
		open,
		cameraName,
		initial,
		submitting = false,
		probing = false,
		deleting = false,
		serverError = null,
		probeResult = null,
		probeError = null,
		onClose,
		onSubmit,
		onProbe,
		onDelete,
	}: Props = $props();

	function handleKeydown(event: KeyboardEvent) {
		if (open && event.key === "Escape") {
			event.preventDefault();
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/70 px-4 py-6 backdrop-blur-sm sm:items-center sm:py-10"
		role="presentation"
	>
		<button
			type="button"
			class="absolute inset-0 cursor-default"
			onclick={onClose}
			aria-label="Close camera settings"
		></button>
		<div
			class="relative z-10 flex max-h-[calc(100vh-3rem)] w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-zinc-700 bg-zinc-950 shadow-2xl sm:max-h-[calc(100vh-5rem)]"
			role="dialog"
			tabindex="-1"
			aria-modal="true"
			aria-labelledby="camera-settings-title"
		>
			<div class="flex shrink-0 items-center justify-between gap-4 border-b border-zinc-800 px-5 py-4 sm:px-6">
				<div class="min-w-0">
					<h2 id="camera-settings-title" class="truncate text-base font-semibold text-zinc-100">
						Camera settings
					</h2>
					<p class="mt-0.5 truncate text-xs text-zinc-500">{cameraName}</p>
				</div>
				<button
					type="button"
					class="rounded-md p-1.5 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-100"
					onclick={onClose}
					aria-label="Close camera settings"
				>
					<span class="text-xl leading-none" aria-hidden="true">×</span>
				</button>
			</div>

			<div class="overflow-y-auto px-5 py-5 sm:px-6">
				<CameraForm
					mode="edit"
					{initial}
					submitting={submitting}
					probing={probing}
					deleting={deleting}
					{serverError}
					{probeResult}
					{probeError}
					onSubmit={onSubmit}
					onProbe={onProbe}
					onDelete={onDelete}
					onCancel={onClose}
				/>
			</div>
		</div>
	</div>
{/if}
