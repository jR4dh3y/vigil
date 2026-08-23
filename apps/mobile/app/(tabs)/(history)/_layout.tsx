import { Stack } from "expo-router/stack";
import { colors } from "@/theme/colors";

export default function HistoryLayout() {
	return (
		<Stack
			screenOptions={{
				contentStyle: { backgroundColor: colors.background },
				headerBackButtonDisplayMode: "minimal",
				headerLargeStyle: { backgroundColor: colors.background },
				headerLargeTitle: true,
				headerShadowVisible: false,
				headerStyle: { backgroundColor: colors.background },
			}}
		>
			<Stack.Screen name="index" options={{ title: "History" }} />
			<Stack.Screen
				name="history/[id]"
				options={{ headerLargeTitle: false, title: "Recordings" }}
			/>
		</Stack>
	);
}
