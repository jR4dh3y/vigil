import type { SystemStatus } from "@nvr/api-client";
import { getApiClient } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";

export async function getSystemStatus(): Promise<SystemStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/system/status");
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not load recorder status");
}

export const systemKeys = {
	all: ["system"] as const,
	status: ["system", "status"] as const,
};
