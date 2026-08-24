import { Color } from "expo-router";
import { Platform } from "react-native";

// Brand palette mirrors apps/landing global.css.
export const colors = {
	paper: "#f3eee3",
	ink: "#1e1e20",
	deepPurple: "#4a3b5f",
	lime: "#c3fb5b",
	coral: "#ff6e79",
	pop: "#ffb86c",
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
	// Brand accent: landing lavender across platforms.
	accent: "#c59edc",
	// Darker semantic shades kept for text on light surfaces where the bright
	// landing hues would fail contrast.
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
	greenSoft: "rgba(195, 251, 91, 0.16)",
	orangeSoft: "rgba(255, 184, 108, 0.2)",
	redSoft: "rgba(255, 110, 121, 0.16)",
	neutralSoft: "rgba(120, 120, 128, 0.14)",
};
