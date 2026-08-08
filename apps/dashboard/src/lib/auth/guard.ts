import { type AuthStatus, isPublicRoute } from "./types";

export type AuthRedirect = { kind: "none" } | { kind: "goto"; to: "/" | "/login" | "/setup" };

/**
 * Decide where the SPA should navigate based on auth status and current path.
 * Pure function — safe to call from effects or derived values.
 */
export function resolveAuthRedirect(
	status: AuthStatus | undefined,
	pathname: string,
): AuthRedirect {
	if (!status) {
		return { kind: "none" };
	}

	if (status.setupRequired) {
		if (pathname !== "/setup") {
			return { kind: "goto", to: "/setup" };
		}
		return { kind: "none" };
	}

	// Setup already done — leave /setup
	if (pathname === "/setup") {
		return { kind: "goto", to: status.user ? "/" : "/login" };
	}

	if (!status.user) {
		if (!isPublicRoute(pathname)) {
			return { kind: "goto", to: "/login" };
		}
		return { kind: "none" };
	}

	// Authenticated users should not stay on login
	if (pathname === "/login") {
		return { kind: "goto", to: "/" };
	}

	return { kind: "none" };
}
