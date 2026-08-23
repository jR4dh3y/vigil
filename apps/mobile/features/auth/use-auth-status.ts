import { useQuery } from "@tanstack/react-query";
import { loadAuthStatus } from "@/features/auth/api";
import { authKeys } from "@/features/auth/keys";

export function useAuthStatus(enabled = true) {
	return useQuery({
		queryKey: authKeys.status,
		queryFn: loadAuthStatus,
		enabled,
		staleTime: 15_000,
		retry: 1,
	});
}
