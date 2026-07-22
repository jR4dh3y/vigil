/**
 * Append `token` as a query param when the URL does not already include one.
 * Preserves relative URLs as relative paths.
 */
export function withStreamToken(url: string, token: string): string {
	if (!token || url.includes("token=")) {
		return url;
	}

	const isAbsolute = /^https?:\/\//i.test(url);
	const base = typeof window !== "undefined" ? window.location.origin : "http://localhost";

	try {
		const parsed = new URL(url, base);
		if (!parsed.searchParams.has("token")) {
			parsed.searchParams.set("token", token);
		}
		if (isAbsolute) {
			return parsed.toString();
		}
		return `${parsed.pathname}${parsed.search}${parsed.hash}`;
	} catch {
		const sep = url.includes("?") ? "&" : "?";
		return `${url}${sep}token=${encodeURIComponent(token)}`;
	}
}

/** Milliseconds until `expiresAt`, or null if unparsable. */
export function msUntilExpiry(expiresAt: string): number | null {
	const t = Date.parse(expiresAt);
	if (Number.isNaN(t)) {
		return null;
	}
	return t - Date.now();
}

/**
 * Query refetch interval that refreshes shortly before the stream token expires.
 * Returns `false` to disable when expiry is unknown.
 */
export function liveRefetchInterval(expiresAt: string | undefined): number | false {
	if (!expiresAt) {
		return false;
	}
	const remaining = msUntilExpiry(expiresAt);
	if (remaining === null) {
		return false;
	}
	// Refresh ~10s before expiry; never poll faster than every 5s.
	return Math.max(remaining - 10_000, 5_000);
}
