import { describe, expect, test } from "bun:test";
import { hydrateStoredPreferences } from "@/lib/preferences";

describe("hydrateStoredPreferences", () => {
	test("preserves the current foreground-alert watermark", () => {
		expect(
			hydrateStoredPreferences({
				notificationsEnabled: true,
				lastSeenEventAt: "2026-08-23T10:00:00Z",
			}),
		).toEqual({
			preferences: {
				notificationsEnabled: true,
				lastSeenEventAt: "2026-08-23T10:00:00Z",
			},
			needsWrite: false,
		});
	});

	test("resets the legacy created-at watermark and carries disarmed intent", () => {
		expect(
			hydrateStoredPreferences({
				armed: false,
				notificationsEnabled: true,
				lastSeenEventAt: "2026-08-23T10:00:00Z",
			}),
		).toEqual({
			preferences: { notificationsEnabled: false, lastSeenEventAt: null },
			needsWrite: true,
		});
	});

	test("rejects malformed preferences", () => {
		expect(
			hydrateStoredPreferences({ notificationsEnabled: "yes", lastSeenEventAt: null }),
		).toBeNull();
		expect(
			hydrateStoredPreferences({
				armed: "yes",
				notificationsEnabled: true,
				lastSeenEventAt: null,
			}),
		).toBeNull();
	});
});
