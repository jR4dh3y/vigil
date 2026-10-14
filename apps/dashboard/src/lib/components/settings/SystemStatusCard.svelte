<script lang="ts">
	import { formatBytes, type SystemStatus } from "$lib/system";

	type Props = {
		status: SystemStatus;
	};

	let { status }: Props = $props();
</script>

<div>
	<div class="flex flex-wrap items-baseline gap-x-5 gap-y-2 text-xs text-zinc-400">
		<p>Version {status.version}</p>
		<p>{status.cameras.online} / {status.cameras.total} cameras online</p>
		<p>{status.retentionDays} day{status.retentionDays === 1 ? "" : "s"} retention</p>
		<p>{formatBytes(status.disk.freeBytes)} of {formatBytes(status.disk.totalBytes)} free</p>
	</div>
	{#if status.health.status !== "ok"}
		<p class="mt-2 text-xs text-amber-300" role="status">Recorder health is degraded.</p>
	{/if}
</div>
