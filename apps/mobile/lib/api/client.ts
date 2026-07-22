import { createApiClient } from "@nvr/api-client";
import { apiBaseUrl } from "@/lib/api/config";

export const api = createApiClient(apiBaseUrl, { credentials: "include" });
