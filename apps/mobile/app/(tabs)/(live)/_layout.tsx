import { Stack } from "expo-router/stack";
import { colors } from "@/theme/colors";

export default function LiveLayout() {
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
			<Stack.Screen name="index" options={{ title: "Live" }} />
			<Stack.Screen name="camera/[id]" options={{ headerLargeTitle: false, title: "Camera" }} />
		</Stack>
	);
}
