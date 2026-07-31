/** Camera detail routes replace the Cameras section tabs with a local context bar. */
export function isCameraDetailRoute(pathname: string): boolean {
	return pathname.startsWith("/cameras/") && pathname !== "/cameras/new" && pathname !== "/cameras";
}
