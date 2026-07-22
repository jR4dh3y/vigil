<script lang="ts">
	import type { EventSeverity } from "$lib/events";
	import { severityBadgeClass, severityLabel } from "$lib/events";

	type Props = {
		severity: EventSeverity;
	};

	let { severity }: Props = $props();
</script>

<span
	class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium {severityBadgeClass(
		severity,
	)}"
>
	{severityLabel(severity)}
</span>
