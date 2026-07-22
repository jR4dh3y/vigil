import { z } from "zod";

const optionalTrimmed = z
	.string()
	.trim()
	.optional()
	.transform((value) => (value && value.length > 0 ? value : undefined));

/** Shared fields for create + edit camera forms. */
const cameraFormBaseSchema = z.object({
	name: z.string().trim().min(1, "Name is required"),
	host: z.string().trim().min(1, "Host is required"),
	username: z.string().trim().optional(),
	password: z.string().optional(),
	enabled: z.boolean(),
	liveRtspUrl: z.string().trim().optional(),
	recordRtspUrl: z.string().trim().optional(),
});

export const createCameraFormSchema = cameraFormBaseSchema;

export const editCameraFormSchema = cameraFormBaseSchema;

export type CreateCameraFormValues = z.infer<typeof createCameraFormSchema>;
export type EditCameraFormValues = z.infer<typeof editCameraFormSchema>;

export const probeFormSchema = z.object({
	rtspUrl: z.string().trim().min(1, "RTSP URL is required to probe"),
	username: optionalTrimmed,
	password: optionalTrimmed,
});

export type ProbeFormValues = z.infer<typeof probeFormSchema>;

/** Flatten Zod field errors to a simple record for form UI. */
export function fieldErrorsFromZod(error: z.ZodError): Record<string, string> {
	const out: Record<string, string> = {};
	for (const issue of error.issues) {
		const key = issue.path[0];
		if (typeof key === "string" && out[key] === undefined) {
			out[key] = issue.message;
		}
	}
	return out;
}
