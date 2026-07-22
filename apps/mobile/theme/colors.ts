import { Color } from "expo-router";
import { Platform } from "react-native";

export const colors = {
	background: Platform.select({
		ios: Color.ios.systemGroupedBackground,
		android: Color.android.dynamic.surface,
		default: "#f4f4f1",
	}),
	surface: Platform.select({
		ios: Color.ios.secondarySystemGroupedBackground,
		android: Color.android.dynamic.surfaceContainer,
		default: "#ffffff",
	}),
	surfaceRaised: Platform.select({
		ios: Color.ios.tertiarySystemGroupedBackground,
		android: Color.android.dynamic.surfaceContainerHigh,
		default: "#ffffff",
	}),
	label: Platform.select({
		ios: Color.ios.label,
		android: Color.android.dynamic.onSurface,
		default: "#171714",
	}),
	secondaryLabel: Platform.select({
		ios: Color.ios.secondaryLabel,
		android: Color.android.dynamic.onSurfaceVariant,
		default: "#686861",
	}),
	separator: Platform.select({
		ios: Color.ios.separator,
		android: Color.android.dynamic.outlineVariant,
		default: "#deded8",
	}),
	accent: Platform.select({
		ios: Color.ios.systemOrange,
		android: Color.android.dynamic.primary,
		default: "#e96f16",
	}),
	green: Platform.select({
		ios: Color.ios.systemGreen,
		android: "#3f8f4d",
		default: "#248a3d",
	}),
	red: Platform.select({
		ios: Color.ios.systemRed,
		android: Color.android.dynamic.error,
		default: "#d73035",
	}),
	orange: Platform.select({
		ios: Color.ios.systemOrange,
		android: "#d97706",
		default: "#e96f16",
	}),
};

export const swatches = {
	preview: "#151817",
	previewRaised: "#232826",
	white: "#ffffff",
	black: "#000000",
	greenSoft: "rgba(53, 199, 89, 0.14)",
	orangeSoft: "rgba(255, 149, 0, 0.14)",
	redSoft: "rgba(255, 59, 48, 0.14)",
	neutralSoft: "rgba(120, 120, 128, 0.14)",
};
