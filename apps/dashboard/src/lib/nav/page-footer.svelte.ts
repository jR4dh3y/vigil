import { createContext, type Snippet } from "svelte";

const [requirePageFooter, setPageFooter] = createContext<{
	content: Snippet | undefined;
}>();

export { setPageFooter };

/** Returns the shell footer context, or undefined when rendered without a shell. */
export function getPageFooter() {
	try {
		return requirePageFooter();
	} catch {
		// Protected pages can render without the shell during an auth redirect.
		return undefined;
	}
}
