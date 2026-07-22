export class ApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "ApiError";
		this.status = status;
		this.code = code;
	}
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

export function throwApiError(body: unknown, status: number, fallback: string): never {
	if (!isRecord(body)) {
		throw new ApiError(fallback, status);
	}

	const message = typeof body.error === "string" && body.error.trim() ? body.error : fallback;
	const code = typeof body.code === "string" ? body.code : undefined;
	throw new ApiError(message, status, code);
}
