<script lang="ts">
	import type { SystemStatus } from "$lib/system";
	import { diskBarClass, diskUsageLabel, formatBytes, healthBadgeClass } from "$lib/system";

	type Props = {
		status: SystemStatus;
	};

	let { status }: Props = $props();

	const usedPercent = $derived(
		Math.min(100, Math.max(0, status.disk.usedPercent)),
	);
</script>

<div class="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 sm:p-6">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div class="flex flex-col gap-1">
			<h2 class="text-sm font-semibold text-zinc-100">System status</h2>
			<p class="text-xs text-zinc-500">Runtime health and storage overview.</p>
		</div>
		<span
			class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize {healthBadgeClass(
				status.health.status,
			)}"
		>
			{status.health.status}
		</span>
	</div>

	<dl class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<div class="flex flex-col gap-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Version</dt>
			<dd class="font-mono text-sm text-zinc-200">{status.version}</dd>
			{#if status.commit}
				<dd class="truncate font-mono text-xs text-zinc-600" title={status.commit}>
					{status.commit.slice(0, 12)}
				</dd>
			{/if}
		</div>

		<div class="flex flex-col gap-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Cameras</dt>
			<dd class="text-sm text-zinc-200">
				<span class="font-medium text-emerald-300">{status.cameras.online}</span>
				<span class="text-zinc-500"> online</span>
			</dd>
			<dd class="text-xs text-zinc-500">
				{status.cameras.enabled} enabled · {status.cameras.total} total
			</dd>
		</div>

		<div class="flex flex-col gap-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Retention</dt>
			<dd class="text-sm text-zinc-200">
				{status.retentionDays} day{status.retentionDays === 1 ? "" : "s"}
			</dd>
		</div>

		<div class="flex flex-col gap-1 sm:col-span-2 lg:col-span-1">
			<dt class="text-xs font-medium tracking-wide text-zinc-500 uppercase">Disk</dt>
			<dd class="text-sm text-zinc-200">{diskUsageLabel(status.disk)}</dd>
			<dd class="text-xs text-zinc-500">
				{usedPercent.toFixed(1)}% used · {formatBytes(status.disk.freeBytes)} free
			</dd>
		</div>
	</dl>

	<div class="mt-5 flex flex-col gap-2">
		<div class="flex items-center justify-between text-xs text-zinc-500">
			<span class="truncate font-mono" title={status.disk.path}>{status.disk.path}</span>
			<span>{usedPercent.toFixed(1)}%</span>
		</div>
		<div class="h-2 overflow-hidden rounded-full bg-zinc-800">
			<div
				class="h-full rounded-full transition-all {diskBarClass(usedPercent)}"
				style="width: {usedPercent}%"
				role="progressbar"
				aria-valuenow={Math.round(usedPercent)}
				aria-valuemin={0}
				aria-valuemax={100}
				aria-label="Disk usage"
			></div>
		</div>
	</div>
</div>
