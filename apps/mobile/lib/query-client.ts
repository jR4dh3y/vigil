import { focusManager, QueryClient } from "@tanstack/react-query";
import { AppState } from "react-native";

export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 30_000,
			retry: 1,
		},
	},
});

focusManager.setEventListener((setFocused) => {
	const subscription = AppState.addEventListener("change", (status) => {
		setFocused(status === "active");
	});
	return () => subscription.remove();
});
