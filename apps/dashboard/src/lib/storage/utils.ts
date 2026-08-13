import type { GDriveArchiveResult, GDriveStatus } from "./types";

export type GDriveCallbackNotice =
	| { kind: "connected"; message: string }
	| { kind: "error"; message: string };

/** Badge styles for Drive connection state (configured / connected / neither). */
export function driveConnectionBadgeClass(
	status: Pick<GDriveStatus, "configured" | "connected" | "connectionError">,
): string {
	if (status.connectionError) {
		return "border-red-500/30 bg-red-500/10 text-red-300";
	}
	if (!status.configured) {
		return "border-zinc-600/40 bg-zinc-800/80 text-zinc-400";
	}
	if (status.connected) {
		return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300";
	}
	return "border-amber-500/30 bg-amber-500/10 text-amber-300";
}

export function driveConnectionLabel(
	status: Pick<GDriveStatus, "configured" | "connected" | "connectionError">,
): string {
	if (status.connectionError) {
		return "Reconnect required";
	}
	if (!status.configured) {
		return "Not configured";
	}
	if (status.connected) {
		return "Connected";
	}
	return "Disconnected";
}

export function formatConnectedAt(value: string | Date | undefined): string | null {
	if (value == null || value === "") {
		return null;
	}
	const date = value instanceof Date ? value : new Date(value);
	if (Number.isNaN(date.getTime())) {
		return typeof value === "string" ? value : null;
	}
	return date.toLocaleString(undefined, {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	});
}

/** Success flash for an archive run, including local cleanup status. */
export function formatGDriveArchiveResult(result: GDriveArchiveResult): string {
	return `Uploaded ${result.uploaded}, deleted local ${result.deleted}, cleanup failed ${result.deleteFailed}, skipped ${result.skipped}, failed ${result.failed}`;
}

/** Reads and sanitizes the one-time OAuth callback result from the settings URL. */
export function readGDriveCallback(params: URLSearchParams): GDriveCallbackNotice | null {
	const result = params.get("gdrive");
	if (result === "connected") {
		return { kind: "connected", message: "Google Drive connected successfully." };
	}
	if (result !== "error") {
		return null;
	}
	const detail = params.get("message")?.trim();
	return {
		kind: "error",
		message: detail
			? `Google Drive connection failed: ${detail}`
			: "Google Drive connection failed. Please try again.",
	};
}

/** Removes OAuth callback parameters without retaining untrusted provider text. */
export function stripGDriveCallback(url: URL): string {
	const next = new URL(url);
	next.searchParams.delete("gdrive");
	next.searchParams.delete("message");
	return `${next.pathname}${next.search}${next.hash}`;
}
