import type { Camera as ApiCamera } from "@nvr/api-client";

export type {
	Camera,
	CreateCameraRequest,
	DiscoverCameraStreamsRequest,
	DiscoverCameraStreamsResult,
	DiscoverCamerasRequest,
	DiscoveredCamera,
	DiscoverResult,
	ProbeCameraRequest,
	ProbeResult,
	StreamProfile,
	UpdateCameraRequest,
} from "@nvr/api-client";

export type CameraStatus = ApiCamera["status"];
