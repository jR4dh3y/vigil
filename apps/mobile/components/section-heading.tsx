import { StyleSheet, Text, View } from "react-native";
import { colors } from "@/theme/colors";

type SectionHeadingProps = {
	title: string;
	detail?: string;
};

export function SectionHeading({ title, detail }: SectionHeadingProps) {
	return (
		<View style={styles.container}>
			<Text selectable style={styles.title}>
				{title}
			</Text>
			{detail ? (
				<Text selectable style={styles.detail}>
					{detail}
				</Text>
			) : null}
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		flexDirection: "row",
		alignItems: "baseline",
		justifyContent: "space-between",
		gap: 12,
	},
	title: {
		color: colors.label,
		fontSize: 18,
		fontWeight: "700",
		letterSpacing: -0.3,
	},
	detail: {
		color: colors.secondaryLabel,
		fontSize: 13,
	},
});
