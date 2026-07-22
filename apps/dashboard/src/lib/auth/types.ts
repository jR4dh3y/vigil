export type {
	ApiErrorBody,
	AuthStatus,
	LoginRequest,
	SetupRequest,
	UserPublic,
} from "@nvr/api-client";

/** Routes that do not require an authenticated session. */
export const PUBLIC_ROUTES = ["/login", "/setup"] as const;

export type PublicRoute = (typeof PUBLIC_ROUTES)[number];

export function isPublicRoute(pathname: string): boolean {
	return (PUBLIC_ROUTES as readonly string[]).includes(pathname);
}
