<script lang="ts">
	import type { RangePreset } from "$lib/recordings";
	import { RANGE_PRESET_LABELS, RANGE_PRESETS } from "$lib/recordings";

	type Props = {
		value: RangePreset;
		disabled?: boolean;
		onChange: (preset: RangePreset) => void;
	};

	let { value, disabled = false, onChange }: Props = $props();
</script>

<div
	class="inline-flex rounded-lg border border-zinc-800 bg-zinc-900/60 p-0.5"
	role="group"
	aria-label="Time range"
>
	{#each RANGE_PRESETS as preset (preset)}
		<button
			type="button"
			class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors
				{value === preset
				? 'bg-zinc-800 text-zinc-100'
				: 'text-zinc-500 hover:text-zinc-300'}
				disabled:cursor-not-allowed disabled:opacity-50"
			disabled={disabled}
			aria-pressed={value === preset}
			onclick={() => onChange(preset)}
		>
			{RANGE_PRESET_LABELS[preset]}
		</button>
	{/each}
</div>
