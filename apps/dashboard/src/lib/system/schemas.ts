import { z } from "zod";

export const settingsFormSchema = z
	.object({
		siteName: z.string().trim().min(1, "Site name is required").max(120, "Site name is too long"),
		retentionDays: z.coerce
			.number()
			.int("Retention must be a whole number")
			.min(1, "Retention must be at least 1 day")
			.max(3650, "Retention cannot exceed 3650 days"),
		recordingsDir: z.string().trim().max(1024, "Path is too long"),
		recordingEnabled: z.boolean(),
	})
	.superRefine((value, ctx) => {
		if (value.recordingEnabled && value.recordingsDir.length === 0) {
			ctx.addIssue({
				code: "custom",
				path: ["recordingsDir"],
				message: "Recording location is required when recording is on",
			});
		}
	});

export type SettingsFormValues = z.infer<typeof settingsFormSchema>;

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
