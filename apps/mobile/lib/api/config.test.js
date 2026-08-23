import { describe, expect, test } from "bun:test";
import {
	createInitialApiConfiguration,
	MediaUrlError,
	normalizeApiBaseUrl,
	resolveMediaUrlAgainstBase,
} from "@/lib/api/config-values";

describe("createInitialApiConfiguration", () => {
	test("keeps the fallback URL explicitly unconfigured", () => {
		expect(createInitialApiConfiguration()).toEqual({
			kind: "unconfigured",
			baseUrl: "http://127.0.0.1:8080/api/v1",
		});
	});

	test("treats a bundled recorder URL as configured", () => {
		expect(createInitialApiConfiguration("https://vigil.example.com")).toEqual({
			kind: "configured",
			baseUrl: "https://vigil.example.com/api/v1",
		});
	});
});

describe("normalizeApiBaseUrl", () => {
	test("adds the API path to a recorder address", () => {
		expect(normalizeApiBaseUrl("192.168.1.8:8080")).toBe("http://192.168.1.8:8080/api/v1");
		expect(normalizeApiBaseUrl("recorder:8080")).toBe("http://recorder:8080/api/v1");
		expect(normalizeApiBaseUrl("vigil.local:8080")).toBe("http://vigil.local:8080/api/v1");
		expect(normalizeApiBaseUrl("vigil.tailnet-name.ts.net:8080")).toBe(
			"http://vigil.tailnet-name.ts.net:8080/api/v1",
		);
	});

	test("rejects credentials and removes query parameters and fragments", () => {
		expect(() => normalizeApiBaseUrl("https://admin:secret@vigil.example.com")).toThrow(
			"must not contain credentials",
		);
		expect(normalizeApiBaseUrl("https://vigil.example.com/?source=app#setup")).toBe(
			"https://vigil.example.com/api/v1",
		);
	});

	test("rejects an explicit unsupported scheme", () => {
		expect(() => normalizeApiBaseUrl("ftp://vigil.example.com")).toThrow("must use HTTP or HTTPS");
	});
});

describe("resolveMediaUrlAgainstBase", () => {
	test("resolves relative media paths against the recorder", () => {
		expect(
			resolveMediaUrlAgainstBase(
				"/live/camera/index.m3u8",
				"http://vigil.lan:8080/api/v1",
				"stream-token",
			),
		).toBe("http://vigil.lan:8080/live/camera/index.m3u8?token=stream-token");
	});

	test("preserves reachable absolute media addresses", () => {
		expect(
			resolveMediaUrlAgainstBase(
				"http://192.168.1.20:8888/camera/whep",
				"http://vigil.lan:8080/api/v1",
			),
		).toBe("http://192.168.1.20:8888/camera/whep");
	});

	test("rewrites HTTP loopback media hosts to the selected recorder host", () => {
		expect(
			resolveMediaUrlAgainstBase(
				"http://127.0.0.2:8888/camera/index.m3u8?quality=high",
				"http://100.101.102.103:8080/api/v1",
				"stream-token",
			),
		).toBe("http://100.101.102.103:8888/camera/index.m3u8?quality=high&token=stream-token");
	});

	test("rejects loopback media advertising through an HTTPS recorder", () => {
		expect(() =>
			resolveMediaUrlAgainstBase(
				"http://localhost:8888/camera/index.m3u8",
				"https://vigil.example.com/api/v1",
			),
		).toThrow(MediaUrlError);
	});
});
