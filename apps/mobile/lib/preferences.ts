export type AppPreferences = {
	notificationsEnabled: boolean;
	lastSeenEventAt: string | null;
};

export type HydratedPreferences = {
	preferences: AppPreferences;
	needsWrite: boolean;
};

/** Validate stored preferences and reset the legacy created-at event watermark. */
export function hydrateStoredPreferences(value: unknown): HydratedPreferences | null {
	if (typeof value !== "object" || value === null) {
		return null;
	}
	if (
		!("notificationsEnabled" in value) ||
		typeof value.notificationsEnabled !== "boolean" ||
		!("lastSeenEventAt" in value) ||
		(value.lastSeenEventAt !== null && typeof value.lastSeenEventAt !== "string")
	) {
		return null;
	}

	if ("armed" in value) {
		if (typeof value.armed !== "boolean") {
			return null;
		}
		return {
			preferences: {
				notificationsEnabled: value.notificationsEnabled && value.armed,
				lastSeenEventAt: null,
			},
			needsWrite: true,
		};
	}

	return {
		preferences: {
			notificationsEnabled: value.notificationsEnabled,
			lastSeenEventAt: value.lastSeenEventAt,
		},
		needsWrite: false,
	};
}
