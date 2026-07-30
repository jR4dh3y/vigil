import { useState } from "react";
import { requestNotificationPermission } from "@/features/notifications/service";
import { useAppStore } from "@/lib/store";

export function useNotificationPreference() {
	const enabled = useAppStore((state) => state.notificationsEnabled);
	const setEnabled = useAppStore((state) => state.setNotificationsEnabled);
	const [requesting, setRequesting] = useState(false);

	const update = async (next: boolean): Promise<boolean> => {
		if (!next) {
			setEnabled(false);
			return true;
		}

		setRequesting(true);
		try {
			const granted = await requestNotificationPermission();
			setEnabled(granted);
			return granted;
		} catch {
			setEnabled(false);
			return false;
		} finally {
			setRequesting(false);
		}
	};

	return { enabled, requesting, update };
}
