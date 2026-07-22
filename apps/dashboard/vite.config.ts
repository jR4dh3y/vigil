import adapter from "@sveltejs/adapter-static";
import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes("node_modules") ? undefined : true,
			},
			// SPA embedded into the Go binary (go:embed). fallback = index.html.
			adapter: adapter({
				pages: "build",
				assets: "build",
				fallback: "index.html",
				precompress: false,
				strict: true,
			}),
		}),
	],
	server: {
		port: 5173,
		strictPort: true,
		proxy: {
			// Same-origin cookies in dev: browser talks to Vite, API to Go on :8080
			"/api": {
				target: "http://127.0.0.1:8080",
				changeOrigin: true,
			},
			// MediaMTX HLS (same-origin avoids Secure cookie / CORS pain on http://localhost)
			"/mtx-hls": {
				target: "http://127.0.0.1:8888",
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/mtx-hls/, ""),
			},
			// MediaMTX WHEP signaling (ICE still goes to MediaMTX UDP :8189)
			"/mtx-webrtc": {
				target: "http://127.0.0.1:8889",
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/mtx-webrtc/, ""),
			},
		},
	},
});
