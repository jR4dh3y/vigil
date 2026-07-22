<script lang="ts">
	import { Camera, Plus } from "lucide-svelte";
</script>

<div
	class="flex min-h-[360px] flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-zinc-800 bg-zinc-900/40 px-6 py-16 text-center"
>
	<span
		class="flex size-12 items-center justify-center rounded-xl bg-zinc-800/80 text-zinc-400"
	>
		<Camera class="size-6" />
	</span>
	<div class="flex flex-col gap-1">
		<p class="text-sm font-medium text-zinc-300">No cameras yet</p>
		<p class="max-w-sm text-sm text-zinc-500">
			Add your first camera with RTSP credentials to start live viewing and recording.
		</p>
	</div>
	<a
		href="/cameras/new"
		class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-sm font-medium text-white no-underline transition-colors hover:bg-emerald-500"
	>
		<Plus class="size-4" />
		Add camera
	</a>
</div>
