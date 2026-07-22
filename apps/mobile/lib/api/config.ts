const DEFAULT_API_URL = "http://127.0.0.1:8080/api/v1";

export const apiBaseUrl = (process.env.EXPO_PUBLIC_API_URL ?? DEFAULT_API_URL).replace(/\/$/, "");

export function resolveMediaUrl(path: string, token?: string): string {
	const baseOrigin = new URL(apiBaseUrl).origin;
	const url = new URL(path, `${baseOrigin}/`);
	if (token) {
		url.searchParams.set("token", token);
	}
	return url.toString();
}
