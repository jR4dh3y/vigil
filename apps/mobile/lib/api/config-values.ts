const DEFAULT_API_URL = "http://127.0.0.1:8080/api/v1";

export type ApiConfiguration =
	| { kind: "configured"; baseUrl: string }
	| { kind: "unconfigured"; baseUrl: string };

export class MediaUrlError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "MediaUrlError";
	}
}

export function normalizeApiBaseUrl(value: string): string {
	const candidate = value.trim();
	if (!candidate) {
		throw new Error("Enter the address of your recorder");
	}
	if (/^[a-z][a-z\d+.-]*:\/\//i.test(candidate) && !/^https?:\/\//i.test(candidate)) {
		throw new Error("Recorder address must use HTTP or HTTPS");
	}

	const withProtocol = /^https?:\/\//i.test(candidate) ? candidate : `http://${candidate}`;
	const url = new URL(withProtocol);
	if (url.protocol !== "http:" && url.protocol !== "https:") {
		throw new Error("Recorder address must use HTTP or HTTPS");
	}
	if (url.username || url.password) {
		throw new Error("Recorder address must not contain credentials");
	}

	url.search = "";
	url.hash = "";
	const path = url.pathname.replace(/\/+$/, "");
	url.pathname = path.endsWith("/api/v1") ? path : `${path}/api/v1`;
	return url.toString().replace(/\/$/, "");
}

export function createInitialApiConfiguration(value?: string): ApiConfiguration {
	const candidate = value?.trim();
	return candidate
		? { kind: "configured", baseUrl: normalizeApiBaseUrl(candidate) }
		: { kind: "unconfigured", baseUrl: DEFAULT_API_URL };
}

export function resolveMediaUrlAgainstBase(
	path: string,
	apiBaseUrl: string,
	token?: string,
): string {
	const recorderUrl = new URL(apiBaseUrl);
	const url = new URL(path, `${recorderUrl.origin}/`);
	if (isLoopbackHostname(url.hostname) && !isLoopbackHostname(recorderUrl.hostname)) {
		if (url.protocol !== "http:" || recorderUrl.protocol !== "http:") {
			throw new MediaUrlError(
				"The recorder advertised a loopback media address that cannot be reached through HTTPS. Configure its HLS, WHEP, and playback URLs with an address this device can reach.",
			);
		}
		url.hostname = recorderUrl.hostname;
	}
	if (token) {
		url.searchParams.set("token", token);
	}
	return url.toString();
}

function isLoopbackHostname(hostname: string): boolean {
	const normalized = hostname.toLowerCase().replace(/^\[|\]$/g, "");
	if (normalized === "localhost" || normalized.endsWith(".localhost") || normalized === "::1") {
		return true;
	}
	const octets = normalized.split(".");
	return octets.length === 4 && octets[0] === "127" && octets.every(isIpv4Octet);
}

function isIpv4Octet(value: string): boolean {
	if (!/^\d{1,3}$/.test(value)) {
		return false;
	}
	const number = Number(value);
	return number >= 0 && number <= 255;
}
