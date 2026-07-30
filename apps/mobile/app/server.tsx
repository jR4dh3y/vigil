import { Stack } from "expo-router";
import { ServerConnectionScreen } from "@/features/server/components/server-connection-screen";
import { useApiBaseUrl } from "@/features/server/use-api-base-url";
import { useServerConnection } from "@/features/server/use-server-connection";

export default function ServerScreen() {
	const baseUrl = useApiBaseUrl();
	const connection = useServerConnection(baseUrl);

	return (
		<>
			<Stack.Title>Recorder</Stack.Title>
			<ServerConnectionScreen
				connecting={connection.connecting}
				error={connection.error}
				onChange={connection.setValue}
				onConnect={connection.connect}
				value={connection.value}
			/>
		</>
	);
}
