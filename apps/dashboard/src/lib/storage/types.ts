export type {
	GDriveArchiveRequest,
	GDriveArchiveResponse,
	GDriveConfigurationRequest,
	GDriveConnectResponse,
	GDriveStatus,
} from "@nvr/api-client";

/** Alias used by the dashboard for archive batch results. */
export type GDriveArchiveResult = import("@nvr/api-client").GDriveArchiveResponse;
