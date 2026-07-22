import { QueryClient } from "@tanstack/svelte-query";
import { browser } from "$app/environment";

export function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser,
				staleTime: 30_000,
				retry: 1,
			},
		},
	});
}
