function parseStreamUrl(url: string): { parsed: URL; isAbsolute: boolean } | null {
	const isAbsolute = /^https?:\/\//i.test(url);
	const base = typeof window !== "undefined" ? window.location.origin : "http://localhost";

	try {
		return { parsed: new URL(url, base), isAbsolute };
	} catch {
		return null;
	}
}

/** Set the current stream token while preserving relative URLs as relative paths. */
export function withStreamToken(url: string, token: string): string {
	if (!token) {
		return url;
	}

	const result = parseStreamUrl(url);
	if (!result) {
		const sep = url.includes("?") ? "&" : "?";
		return `${url}${sep}token=${encodeURIComponent(token)}`;
	}

	result.parsed.searchParams.set("token", token);
	return result.isAbsolute
		? result.parsed.toString()
		: `${result.parsed.pathname}${result.parsed.search}${result.parsed.hash}`;
}

/** Remove the rotating token to get a stable stream endpoint identity. */
export function streamEndpoint(url: string): string {
	const result = parseStreamUrl(url);
	if (!result) {
		return url;
	}

	result.parsed.searchParams.delete("token");
	return result.isAbsolute
		? result.parsed.toString()
		: `${result.parsed.pathname}${result.parsed.search}${result.parsed.hash}`;
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
