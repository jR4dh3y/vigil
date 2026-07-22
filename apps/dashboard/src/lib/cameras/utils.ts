import type {
	Camera,
	CreateCameraRequest,
	StreamProfile,
	UpdateCameraRequest,
} from "@nvr/api-client";
import type { CreateCameraFormValues, EditCameraFormValues } from "./schemas";

const DEFAULT_DRIVER = "generic-rtsp";

/** Prefer live profile codec, then record, then any profile. */
export function primaryCodec(profiles: StreamProfile[]): string | null {
	const live = profiles.find((p) => p.role === "live" && p.codec);
	if (live?.codec) {
		return live.codec;
	}
	const record = profiles.find((p) => p.role === "record" && p.codec);
	if (record?.codec) {
		return record.codec;
	}
	const any = profiles.find((p) => p.codec);
	return any?.codec ?? null;
}

/** Prefer live profile resolution. */
export function primaryResolution(profiles: StreamProfile[]): string | null {
	const ordered = [
		...profiles.filter((p) => p.role === "live"),
		...profiles.filter((p) => p.role === "record"),
		...profiles,
	];
	for (const profile of ordered) {
		if (profile.width && profile.height) {
			return `${profile.width}×${profile.height}`;
		}
	}
	return null;
}

/**
 * Resolve RTSP URL for probe: liveRtspUrl first, else host when it is already an RTSP URL.
 */
export function resolveProbeRtspUrl(
	liveRtspUrl: string | undefined,
	host: string | undefined,
): string | null {
	const live = liveRtspUrl?.trim();
	if (live) {
		return live;
	}
	const h = host?.trim();
	if (!h) {
		return null;
	}
	const lower = h.toLowerCase();
	if (lower.startsWith("rtsp://") || lower.startsWith("rtsps://")) {
		return h;
	}
	return null;
}

export function toCreateCameraRequest(values: CreateCameraFormValues): CreateCameraRequest {
	const body: CreateCameraRequest = {
		name: values.name,
		host: values.host,
		driver: DEFAULT_DRIVER,
		enabled: values.enabled,
	};

	const username = values.username?.trim();
	if (username) {
		body.username = username;
	}
	const password = values.password;
	if (password) {
		body.password = password;
	}
	const liveRtspUrl = values.liveRtspUrl?.trim();
	if (liveRtspUrl) {
		body.liveRtspUrl = liveRtspUrl;
	}
	const recordRtspUrl = values.recordRtspUrl?.trim();
	if (recordRtspUrl) {
		body.recordRtspUrl = recordRtspUrl;
	}

	return body;
}

/**
 * Build PATCH body. Empty password is omitted (leave unchanged).
 * Empty optional strings for username/URLs are sent as empty only when user cleared them intentionally —
 * we omit empty username/URLs so server leaves them unchanged if that is the convention.
 */
export function toUpdateCameraRequest(values: EditCameraFormValues): UpdateCameraRequest {
	const body: UpdateCameraRequest = {
		name: values.name,
		host: values.host,
		enabled: values.enabled,
	};

	const username = values.username?.trim();
	if (username) {
		body.username = username;
	}

	const password = values.password;
	if (password && password.length > 0) {
		body.password = password;
	}

	const liveRtspUrl = values.liveRtspUrl?.trim();
	if (liveRtspUrl) {
		body.liveRtspUrl = liveRtspUrl;
	}

	const recordRtspUrl = values.recordRtspUrl?.trim();
	if (recordRtspUrl) {
		body.recordRtspUrl = recordRtspUrl;
	}

	return body;
}

export function formValuesFromCamera(camera: Camera): EditCameraFormValues {
	const live = camera.streamProfiles.find((p) => p.role === "live");
	const record = camera.streamProfiles.find((p) => p.role === "record");
	return {
		name: camera.name,
		host: camera.host,
		username: "",
		password: "",
		enabled: camera.enabled,
		liveRtspUrl: live?.rtspUrl ?? "",
		recordRtspUrl: record?.rtspUrl ?? "",
	};
}

export function statusLabel(status: Camera["status"]): string {
	switch (status) {
		case "online":
			return "Online";
		case "offline":
			return "Offline";
		default:
			return "Unknown";
	}
}
