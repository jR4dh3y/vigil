import { useMutation, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import { useState } from "react";
import { testRecorderConnection } from "@/features/server/api";
import { saveApiBaseUrl } from "@/lib/api/config";
import { clearSessionToken } from "@/lib/api/session";
import { useAppStore } from "@/lib/store";

export function useServerConnection(initialUrl: string) {
	const queryClient = useQueryClient();
	const [value, setValue] = useState(initialUrl);
	const [error, setError] = useState<string | null>(null);
	const mutation = useMutation({
		mutationFn: () => testRecorderConnection(value),
		onSuccess: async ({ baseUrl }) => {
			await saveApiBaseUrl(baseUrl);
			await clearSessionToken();
			useAppStore.getState().setLastSeenEventAt(null);
			queryClient.clear();
			setError(null);
			router.replace("/");
		},
		onError: (cause: unknown) => {
			setError(cause instanceof Error ? cause.message : "Could not connect to this recorder");
		},
	});

	return {
		value,
		error,
		connecting: mutation.isPending,
		setValue,
		connect: () => {
			setError(null);
			mutation.mutate();
		},
	};
}
