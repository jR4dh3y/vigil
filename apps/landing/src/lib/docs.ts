import { existsSync, readdirSync, readFileSync } from "node:fs";
import { basename, join, resolve } from "node:path";
import { marked } from "marked";
import { codeToHtml } from "shiki";

export type DocNavItem = {
	slug: string;
	title: string;
	description: string;
};

export type DocHeading = {
	depth: number;
	text: string;
	id: string;
};

export type DocPage = DocNavItem & {
	html: string;
	headings: DocHeading[];
};

const docsDirectory = resolveDocsDirectory();

const docMeta: Record<string, Omit<DocNavItem, "slug">> = {
	"getting-started": {
		title: "Getting Started",
		description: "How to run Vigil.",
	},
	architecture: {
		title: "Architecture",
		description: "The system design and data flows.",
	},
	backend: {
		title: "Backend",
		description: "The Go server (nvrd).",
	},
	api: {
		title: "API",
		description: "The REST API and WebSocket.",
	},
	database: {
		title: "Database",
		description: "The SQLite schema.",
	},
	dashboard: {
		title: "Dashboard",
		description: "The web dashboard.",
	},
	mobile: {
		title: "Mobile",
		description: "The mobile app.",
	},
	landing: {
		title: "Landing",
		description: "The marketing website.",
	},
	deployment: {
		title: "Deployment",
		description: "Docker, MediaMTX, and environment variables.",
	},
	tooling: {
		title: "Tooling",
		description: "Build, code generation, and linting.",
	},
	"ste100-style-guide": {
		title: "Style Guide",
		description: "The Simplified Technical English rules.",
	},
};

const navOrder = [
	"getting-started",
	"architecture",
	"backend",
	"api",
	"database",
	"dashboard",
	"mobile",
	"landing",
	"deployment",
	"tooling",
	"ste100-style-guide",
];

marked.setOptions({
	gfm: true,
	breaks: false,
});

export function getDocNav(): DocNavItem[] {
	const available = new Set(getDocSlugs());

	return navOrder
		.filter((slug) => available.has(slug))
		.map((slug) => ({
			slug,
			...(docMeta[slug] ?? {
				title: titleFromSlug(slug),
				description: "",
			}),
		}));
}

export function getDocSlugs(): string[] {
	return readdirSync(docsDirectory)
		.filter((file) => file.endsWith(".md"))
		.map((file) => basename(file, ".md"));
}

export async function getDocPage(slug: string): Promise<DocPage> {
	const normalizedSlug = slug === "" || slug === "index" ? "README" : slug;
	const markdown = readFileSync(join(docsDirectory, `${normalizedSlug}.md`), "utf8");
	const rendered = await marked.parse(markdown);
	const highlighted = await highlightCodeBlocks(rendered);
	const linked = normalizeDocLinks(highlighted);
	const { html, headings } = addHeadingIds(linked);
	const meta = docMeta[normalizedSlug] ?? {
		title: titleFromMarkdown(markdown) ?? titleFromSlug(normalizedSlug),
		description: "",
	};

	return {
		slug: normalizedSlug,
		...meta,
		html,
		headings,
	};
}

function resolveDocsDirectory(): string {
	const directory = resolve(process.cwd(), "../../docs");
	if (!existsSync(directory)) {
		throw new Error("Unable to find local docs directory.");
	}

	return directory;
}

function addHeadingIds(html: string): { html: string; headings: DocHeading[] } {
	const headings: DocHeading[] = [];
	const usedIds = new Map<string, number>();

	const withIds = html.replace(
		/<h([23])>(.*?)<\/h\1>/g,
		(_match, depth: string, content: string) => {
			const text = content.replace(/<[^>]+>/g, "").trim();
			const baseId = slugify(text);
			const count = usedIds.get(baseId) ?? 0;
			const id = count === 0 ? baseId : `${baseId}-${count + 1}`;
			usedIds.set(baseId, count + 1);
			headings.push({ depth: Number(depth), text, id });

			return `<h${depth} id="${id}">${content}</h${depth}>`;
		},
	);

	return { html: withIds, headings };
}

function normalizeDocLinks(html: string): string {
	return html.replace(
		/href="(\.\/)?([^"#]+)\.md(#[^"]*)?"/g,
		(_match, _prefix: string, slug: string, hash = "") => {
			const route = slug === "README" ? "/docs/" : `/docs/${slug}/`;

			return `href="${route}${hash}"`;
		},
	);
}

async function highlightCodeBlocks(html: string): Promise<string> {
	const codeBlockPattern = /<pre><code(?: class="language-([^"]+)")?>([\s\S]*?)<\/code><\/pre>/g;
	const parts: string[] = [];
	let lastIndex = 0;

	for (const match of html.matchAll(codeBlockPattern)) {
		const [block, language, encodedCode] = match;
		const index = match.index ?? 0;
		parts.push(html.slice(lastIndex, index));
		parts.push(await highlightCodeBlock(decodeCodeHtml(encodedCode), language));
		lastIndex = index + block.length;
	}

	parts.push(html.slice(lastIndex));
	return parts.join("");
}

async function highlightCodeBlock(code: string, language?: string): Promise<string> {
	const lang = normalizeCodeLanguage(language);

	try {
		return await codeToHtml(code, {
			lang,
			theme: "github-dark-default",
		});
	} catch {
		return codeToHtml(code, {
			lang: "text",
			theme: "github-dark-default",
		});
	}
}

function normalizeCodeLanguage(language?: string): string {
	if (!language) {
		return "text";
	}

	const aliases: Record<string, string> = {
		compose: "yaml",
		plaintext: "text",
		sh: "bash",
		shell: "bash",
		yml: "yaml",
	};

	return aliases[language] ?? language;
}

function decodeCodeHtml(value: string): string {
	return value
		.replace(/&amp;/g, "&")
		.replace(/&lt;/g, "<")
		.replace(/&gt;/g, ">")
		.replace(/&quot;/g, '"')
		.replace(/&#39;/g, "'");
}

function titleFromMarkdown(markdown: string): string | null {
	const match = markdown.match(/^#\s+(.+)$/m);
	return match?.[1] ?? null;
}

function titleFromSlug(slug: string): string {
	return slug
		.split("-")
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(" ");
}

function slugify(value: string): string {
	return value
		.toLowerCase()
		.replace(/`/g, "")
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "");
}
