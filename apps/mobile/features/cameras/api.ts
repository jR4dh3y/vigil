import type { Camera, LiveStream } from "@nvr/api-client";
import { api } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";

export async function listCameras(): Promise<Camera[]> {
	const { data, error, response } = await api.GET("/cameras");
	if (data) {
		return data.cameras;
	}
	throwApiError(error, response.status, "Could not load cameras");
}

export async function getCamera(id: string): Promise<Camera> {
	const { data, error, response } = await api.GET("/cameras/{id}", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not load camera");
}

export async function getLiveStream(id: string): Promise<LiveStream> {
	const { data, error, response } = await api.POST("/cameras/{id}/live", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Live stream is unavailable");
}

export const cameraKeys = {
	all: ["cameras"] as const,
	detail: (id: string) => ["cameras", id] as const,
	live: (id: string) => ["cameras", id, "live"] as const,
};
