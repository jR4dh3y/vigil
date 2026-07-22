import { z } from "zod";

export const userRoles = ["admin", "operator", "viewer"] as const;

export const createUserFormSchema = z.object({
	username: z.string().trim().min(1, "Username is required").max(64, "Username is too long"),
	password: z.string().min(8, "Password must be at least 8 characters"),
	role: z.enum(userRoles, { error: "Role is required" }),
});

export type CreateUserFormValues = z.infer<typeof createUserFormSchema>;

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
