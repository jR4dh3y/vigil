import createClient from "openapi-fetch";
import type { components, paths } from "./generated/schema";

export type { components, paths };
export type ApiClient = ReturnType<typeof createClient<paths>>;

export type UserPublic = components["schemas"]["UserPublic"];
export type AuthStatus = components["schemas"]["AuthStatus"];
export type LoginRequest = components["schemas"]["LoginRequest"];
export type SetupRequest = components["schemas"]["SetupRequest"];
export type ApiErrorBody = components["schemas"]["Error"];

export type Camera = components["schemas"]["Camera"];
export type StreamProfile = components["schemas"]["StreamProfile"];
export type CreateCameraRequest = components["schemas"]["CreateCameraRequest"];
export type UpdateCameraRequest = components["schemas"]["UpdateCameraRequest"];
export type DiscoveredCamera = components["schemas"]["DiscoveredCamera"];
export type DiscoverResult = components["schemas"]["DiscoverResult"];
export type ProbeCameraRequest = components["schemas"]["ProbeCameraRequest"];
export type ProbeResult = components["schemas"]["ProbeResult"];
export type LiveStream = components["schemas"]["LiveStream"];

export type CoverageBar = components["schemas"]["CoverageBar"];
export type RecordingSegment = components["schemas"]["RecordingSegment"];
export type RecordingList = components["schemas"]["RecordingList"];
export type PlaybackRequest = components["schemas"]["PlaybackRequest"];
export type PlaybackSession = components["schemas"]["PlaybackSession"];

export type Event = components["schemas"]["Event"];
export type EventList = components["schemas"]["EventList"];
export type EventSeverity = Event["severity"];
export type DiskInfo = components["schemas"]["DiskInfo"];
export type SystemStatus = components["schemas"]["SystemStatus"];
export type Settings = components["schemas"]["Settings"];
export type PatchSettingsRequest = components["schemas"]["PatchSettingsRequest"];
export type CreateUserRequest = components["schemas"]["CreateUserRequest"];
export type UserRole = UserPublic["role"];
export type GDriveStatus = components["schemas"]["GDriveStatus"];
export type GDriveConfigurationRequest = components["schemas"]["GDriveConfigurationRequest"];
export type GDriveConnectResponse = components["schemas"]["GDriveConnectResponse"];
export type GDriveArchiveRequest = components["schemas"]["GDriveArchiveRequest"];
export type GDriveArchiveResponse = components["schemas"]["GDriveArchiveResponse"];

export type CreateApiClientOptions = {
	/** Defaults to `"include"` so cookie sessions work cross-origin when needed. */
	credentials?: RequestCredentials;
	/** Optional platform-specific fetch implementation. */
	fetch?: (input: Request) => Promise<Response>;
};

export function createApiClient(baseUrl: string, options?: CreateApiClientOptions): ApiClient {
	return createClient<paths>({
		baseUrl,
		credentials: options?.credentials ?? "include",
		fetch: options?.fetch,
	});
}

export type { Middleware } from "openapi-fetch";
