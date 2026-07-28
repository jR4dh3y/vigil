import { Stack } from "expo-router/stack";
import { colors } from "@/theme/colors";

export default function EventsLayout() {
	return (
		<Stack
			screenOptions={{
				contentStyle: { backgroundColor: colors.background },
				headerLargeStyle: { backgroundColor: colors.background },
				headerLargeTitle: true,
				headerShadowVisible: false,
				headerStyle: { backgroundColor: colors.background },
			}}
		>
			<Stack.Screen name="index" options={{ title: "Events" }} />
			<Stack.Screen name="event/[id]" options={{ headerLargeTitle: false, title: "Event" }} />
		</Stack>
	);
}
