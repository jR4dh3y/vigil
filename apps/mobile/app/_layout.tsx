import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { useState } from "react";

export default function RootLayout() {
	const [queryClient] = useState(
		() =>
			new QueryClient({
				defaultOptions: {
					queries: { staleTime: 30_000, retry: 1 },
				},
			}),
	);

	return (
		<QueryClientProvider client={queryClient}>
			<StatusBar style="auto" />
			<Stack screenOptions={{ headerShown: true, title: "NVR" }} />
		</QueryClientProvider>
	);
}
