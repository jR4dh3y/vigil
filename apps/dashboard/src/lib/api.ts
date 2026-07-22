import { createApiClient } from "@nvr/api-client";

const baseUrl = import.meta.env.VITE_API_BASE ?? "/api/v1";

/** Shared OpenAPI client (cookie sessions via credentials: include). */
export const api = createApiClient(baseUrl, { credentials: "include" });
