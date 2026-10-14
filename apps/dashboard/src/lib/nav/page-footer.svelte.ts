import { createContext, type Snippet } from "svelte";

export const [getPageFooter, setPageFooter] = createContext<{
	content: Snippet | undefined;
}>();
