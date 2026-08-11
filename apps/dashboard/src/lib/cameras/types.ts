import type { Camera as ApiCamera } from "@nvr/api-client";

export type {
	Camera,
	CreateCameraRequest,
	DiscoveredCamera,
	DiscoverResult,
	ProbeCameraRequest,
	ProbeResult,
	StreamProfile,
	UpdateCameraRequest,
} from "@nvr/api-client";

export type CameraStatus = ApiCamera["status"];
