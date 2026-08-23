import { QueryClientProvider } from "@tanstack/react-query";
import { Stack } from "expo-router";
import { DarkTheme, DefaultTheme, ThemeProvider } from "expo-router/react-navigation";
import { StatusBar } from "expo-status-bar";
import { useEffect, useState } from "react";
import { ActivityIndicator, Platform, StyleSheet, useColorScheme, View } from "react-native";
import { useProtectedSessionRouting } from "@/features/auth/use-protected-session-routing";
import { useNotificationRouting } from "@/features/notifications/use-notification-routing";
import { hydrateApiConfig } from "@/lib/api/config";
import { hydrateSession } from "@/lib/api/session";
import { queryClient } from "@/lib/query-client";
import { hydrateAppPreferences } from "@/lib/store";
import { colors } from "@/theme/colors";

export default function RootLayout() {
	const colorScheme = useColorScheme();
	const [sessionReady, setSessionReady] = useState(false);
	useProtectedSessionRouting();
	useNotificationRouting(sessionReady);

	useEffect(() => {
		void Promise.all([hydrateApiConfig(), hydrateSession(), hydrateAppPreferences()]).finally(() =>
			setSessionReady(true),
		);
	}, []);

	if (!sessionReady) {
		return (
			<View style={styles.boot}>
				<ActivityIndicator color={colors.accent} size="large" />
			</View>
		);
	}

	return (
		<QueryClientProvider client={queryClient}>
			<ThemeProvider value={colorScheme === "dark" ? DarkTheme : DefaultTheme}>
				<StatusBar style="auto" />
				<Stack screenOptions={{ headerShown: false }}>
					<Stack.Screen name="(tabs)" />
					<Stack.Screen name="login" options={{ animation: "fade" }} />
					<Stack.Screen name="setup" options={{ animation: "fade" }} />
					<Stack.Screen
						name="server"
						options={{
							headerShown: true,
							presentation: "formSheet",
							sheetAllowedDetents: Platform.OS === "ios" ? [0.75, 1] : "fitToContents",
							sheetGrabberVisible: true,
							sheetCornerRadius: 24,
						}}
					/>
					<Stack.Screen name="+not-found" options={{ headerShown: true, title: "Not found" }} />
				</Stack>
			</ThemeProvider>
		</QueryClientProvider>
	);
}

const styles = StyleSheet.create({
	boot: {
		alignItems: "center",
		backgroundColor: colors.background,
		flex: 1,
		justifyContent: "center",
	},
});
