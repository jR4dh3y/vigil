import { z } from "zod";

export const setupFormSchema = z.object({
	username: z.string().min(1, "Username is required"),
	password: z.string().min(8, "Password must be at least 8 characters"),
});

export const loginFormSchema = z.object({
	username: z.string().min(1, "Username is required"),
	password: z.string().min(1, "Password is required"),
});

export type SetupFormValues = z.infer<typeof setupFormSchema>;
export type LoginFormValues = z.infer<typeof loginFormSchema>;

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
