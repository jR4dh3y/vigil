/** DOM hosts in the top bar where page content is portaled. */
export const PAGE_META_HOST_ID = "nvr-page-meta";
export const PAGE_ACTIONS_HOST_ID = "nvr-page-actions";

export type PageActionsSide = "start" | "end";

export function getPageActionsHost(side: PageActionsSide = "end"): HTMLElement | null {
	if (typeof document === "undefined") {
		return null;
	}
	const id = side === "start" ? PAGE_META_HOST_ID : PAGE_ACTIONS_HOST_ID;
	return document.getElementById(id);
}
