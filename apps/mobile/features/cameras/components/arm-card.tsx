import { Pressable, StyleSheet, Text, View } from "react-native";
import { useAppStore } from "@/lib/store";
import { swatches } from "@/theme/colors";

export function ArmCard() {
	const armed = useAppStore((state) => state.armed);
	const setArmed = useAppStore((state) => state.setArmed);

	return (
		<View style={[styles.container, armed ? styles.armed : styles.disarmed]}>
			<View style={styles.copy}>
				<View style={[styles.icon, armed ? styles.iconArmed : styles.iconDisarmed]}>
					<View style={[styles.iconCore, armed ? styles.iconCoreArmed : styles.iconCoreDisarmed]} />
				</View>
				<View style={styles.textGroup}>
					<Text selectable style={styles.eyebrow}>
						HOME MONITORING
					</Text>
					<Text selectable style={styles.title}>
						{armed ? "Armed and watching" : "Monitoring paused"}
					</Text>
					<Text selectable style={styles.detail}>
						{armed
							? "This device will surface new camera alerts."
							: "Live video is still available."}
					</Text>
				</View>
			</View>
			<Pressable
				accessibilityLabel={armed ? "Disarm home monitoring" : "Arm home monitoring"}
				accessibilityRole="button"
				onPress={() => setArmed(!armed)}
				style={({ pressed }) => [styles.button, pressed && styles.buttonPressed]}
			>
				<Text style={styles.buttonLabel}>{armed ? "Disarm" : "Arm now"}</Text>
			</Pressable>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		borderCurve: "continuous",
		borderRadius: 26,
		gap: 24,
		padding: 22,
	},
	armed: {
		backgroundColor: "#173c2b",
	},
	disarmed: {
		backgroundColor: "#46321f",
	},
	copy: {
		alignItems: "center",
		flexDirection: "row",
		gap: 15,
	},
	icon: {
		alignItems: "center",
		borderRadius: 18,
		height: 54,
		justifyContent: "center",
		width: 54,
	},
	iconArmed: {
		backgroundColor: "rgba(92, 224, 142, 0.15)",
	},
	iconDisarmed: {
		backgroundColor: "rgba(255, 184, 77, 0.16)",
	},
	iconCore: {
		borderRadius: 6,
		borderWidth: 3,
		height: 22,
		transform: [{ rotate: "45deg" }],
		width: 22,
	},
	iconCoreArmed: {
		borderColor: "#6ee7a0",
	},
	iconCoreDisarmed: {
		borderColor: "#ffb84d",
	},
	textGroup: {
		flex: 1,
		gap: 3,
	},
	eyebrow: {
		color: "rgba(255, 255, 255, 0.62)",
		fontSize: 10,
		fontWeight: "800",
		letterSpacing: 1.2,
	},
	title: {
		color: swatches.white,
		fontSize: 19,
		fontWeight: "800",
		letterSpacing: -0.35,
	},
	detail: {
		color: "rgba(255, 255, 255, 0.68)",
		fontSize: 13,
	},
	button: {
		alignItems: "center",
		backgroundColor: swatches.white,
		borderCurve: "continuous",
		borderRadius: 14,
		paddingVertical: 12,
	},
	buttonPressed: {
		opacity: 0.75,
		transform: [{ scale: 0.99 }],
	},
	buttonLabel: {
		color: swatches.black,
		fontSize: 14,
		fontWeight: "800",
	},
});
