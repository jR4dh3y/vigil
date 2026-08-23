import { Stack, useLocalSearchParams } from "expo-router";
import { useIsFocused } from "expo-router/react-navigation";
import { RecordingHistory } from "@/features/recordings/components/recording-history";
import { useRecordingHistory } from "@/features/recordings/use-recording-history";

export default function RecordingHistoryScreen() {
	const { id } = useLocalSearchParams<{ id: string }>();
	const isFocused = useIsFocused();
	const history = useRecordingHistory(id, isFocused);

	return (
		<>
			<Stack.Title>{history.camera?.name ?? "Recordings"}</Stack.Title>
			<RecordingHistory history={history} />
		</>
	);
}
