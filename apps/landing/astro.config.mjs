// @ts-check
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";

// https://astro.build/config
export default defineConfig({
	vite: {
		plugins: [tailwindcss()],
		css: {
			// Astro 7 defaults to lightningcss; PostCSS is required for
			// Tailwind v4's Vite plugin to intercept CSS transforms.
			transformer: "postcss",
		},
	},
});
