import type { GDriveArchiveResult } from "./types";

export type GDriveCallbackNotice =
	| { kind: "connected"; message: string }
	| { kind: "error"; message: string };

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
