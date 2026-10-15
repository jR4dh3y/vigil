import { createContext, type Snippet } from "svelte";

const [requirePageFooter, setPageFooter] = createContext<{
	content: Snippet | undefined;
}>();

export { setPageFooter };

export function getPageFooter() {
	try {
		return requirePageFooter();
	} catch {
		// Protected pages can render without the shell during an auth redirect.
		return undefined;
	}
}
