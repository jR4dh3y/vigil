<script lang="ts">
	import { Activity, CalendarClock, Camera, HardDrive, Server } from "lucide-svelte";
	import type { SystemStatus } from "$lib/system";
	import { diskBarClass, diskUsageLabel, formatBytes, healthBadgeClass } from "$lib/system";

	type Props = {
		status: SystemStatus;
	};

	let { status }: Props = $props();

	const usedPercent = $derived(Math.min(100, Math.max(0, status.disk.usedPercent)));
</script>

<div class="overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/35">
	<div
		class="flex flex-wrap items-center justify-between gap-4 border-b border-zinc-800 px-5 py-4 sm:px-6"
	>
		<div class="flex items-center gap-3">
			<span
				class="flex size-9 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-400"
			>
				<Activity class="size-4" />
			</span>
			<div>
				<h2 class="text-sm font-semibold text-zinc-100">System overview</h2>
				<p class="mt-0.5 text-xs text-zinc-500">Current recorder health and capacity.</p>
			</div>
		</div>
		<span
			class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium capitalize {healthBadgeClass(
				status.health.status,
			)}"
		>
			<span class="size-1.5 rounded-full bg-current"></span>
			{status.health.status}
		</span>
	</div>

	<div class="grid sm:grid-cols-2 lg:grid-cols-4">
		<div class="flex min-w-0 gap-3 border-b border-zinc-800 p-5 sm:border-r lg:border-b-0">
			<span
				class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-zinc-800/70 text-zinc-400"
			>
				<Server class="size-4" />
			</span>
			<div class="min-w-0">
				<p class="text-xs text-zinc-500">Version</p>
				<p class="mt-1 truncate font-mono text-sm font-medium text-zinc-200">
					{status.version}
				</p>
				{#if status.commit}
					<p class="mt-0.5 truncate font-mono text-[11px] text-zinc-600" title={status.commit}>
						{status.commit.slice(0, 12)}
					</p>
				{/if}
			</div>
		</div>

		<div class="flex min-w-0 gap-3 border-b border-zinc-800 p-5 lg:border-r lg:border-b-0">
			<span
				class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-zinc-800/70 text-zinc-400"
			>
				<Camera class="size-4" />
			</span>
			<div class="min-w-0">
				<p class="text-xs text-zinc-500">Cameras</p>
				<p class="mt-1 text-sm font-medium text-zinc-200">
					<span class="text-emerald-300">{status.cameras.online}</span> online
				</p>
				<p class="mt-0.5 text-[11px] text-zinc-500">
					{status.cameras.enabled} enabled · {status.cameras.total} total
				</p>
			</div>
		</div>

		<div class="flex min-w-0 gap-3 border-b border-zinc-800 p-5 sm:border-r lg:border-b-0">
			<span
				class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-zinc-800/70 text-zinc-400"
			>
				<CalendarClock class="size-4" />
			</span>
			<div class="min-w-0">
				<p class="text-xs text-zinc-500">Retention</p>
				<p class="mt-1 text-sm font-medium text-zinc-200">
					{status.retentionDays} day{status.retentionDays === 1 ? "" : "s"}
				</p>
				<p class="mt-0.5 text-[11px] text-zinc-500">Rolling local history</p>
			</div>
		</div>

		<div class="flex min-w-0 gap-3 p-5">
			<span
				class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-zinc-800/70 text-zinc-400"
			>
				<HardDrive class="size-4" />
			</span>
			<div class="min-w-0 flex-1">
				<div class="flex items-center justify-between gap-3">
					<p class="text-xs text-zinc-500">Storage</p>
					<span class="text-[11px] text-zinc-500">{usedPercent.toFixed(1)}%</span>
				</div>
				<p class="mt-1 truncate text-sm font-medium text-zinc-200">
					{formatBytes(status.disk.freeBytes)} free
				</p>
				<div class="mt-2">
					<div class="h-1.5 overflow-hidden rounded-full bg-zinc-800">
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
		</div>
	</div>

	<div
		class="flex flex-col gap-1 border-t border-zinc-800 bg-zinc-950/30 px-5 py-3 text-[11px] text-zinc-500 sm:flex-row sm:items-center sm:justify-between sm:px-6"
	>
		<span class="truncate font-mono" title={status.disk.path}>{status.disk.path}</span>
		<span class="shrink-0">{diskUsageLabel(status.disk)}</span>
	</div>
</div>
