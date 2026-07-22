import { QueryClientProvider } from "@tanstack/react-query";
import { DarkTheme, DefaultTheme, ThemeProvider } from "expo-router/react-navigation";
import { NativeTabs } from "expo-router/unstable-native-tabs";
import { StatusBar } from "expo-status-bar";
import { useColorScheme } from "react-native";
import { queryClient } from "@/lib/query-client";

export default function RootLayout() {
	const colorScheme = useColorScheme();

	return (
		<QueryClientProvider client={queryClient}>
			<ThemeProvider value={colorScheme === "dark" ? DarkTheme : DefaultTheme}>
				<StatusBar style="auto" />
				<NativeTabs>
					<NativeTabs.Trigger name="(live)">
						<NativeTabs.Trigger.Icon sf="rectangle.grid.2x2.fill" md="dashboard" />
						<NativeTabs.Trigger.Label>Live</NativeTabs.Trigger.Label>
					</NativeTabs.Trigger>
					<NativeTabs.Trigger name="(events)">
						<NativeTabs.Trigger.Icon sf="bell.fill" md="notifications" />
						<NativeTabs.Trigger.Label>Events</NativeTabs.Trigger.Label>
					</NativeTabs.Trigger>
					<NativeTabs.Trigger name="(settings)">
						<NativeTabs.Trigger.Icon sf="gearshape.fill" md="settings" />
						<NativeTabs.Trigger.Label>Settings</NativeTabs.Trigger.Label>
					</NativeTabs.Trigger>
				</NativeTabs>
			</ThemeProvider>
		</QueryClientProvider>
	);
}
