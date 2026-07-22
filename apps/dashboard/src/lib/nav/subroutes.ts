export type SubRoute = {
	href: string;
	label: string;
	/** Custom active match; defaults to exact or prefix of href. */
	isActive?: (pathname: string) => boolean;
	/** When true (default false), only match exact path. */
	exact?: boolean;
	/** Hide unless user is admin. */
	adminOnly?: boolean;
};

export type MainSection = {
	id: string;
	/** Prefix used to detect the section from pathname. */
	match: (pathname: string) => boolean;
	title: string;
	subRoutes: SubRoute[];
};

const sections: MainSection[] = [
	{
		id: "live",
		match: (pathname) => pathname === "/",
		title: "Live",
		subRoutes: [],
	},
	{
		id: "cameras",
		match: (pathname) => pathname === "/cameras" || pathname.startsWith("/cameras/"),
		title: "Cameras",
		subRoutes: [
			{
				href: "/cameras",
				label: "All",
				isActive: (pathname) =>
					pathname === "/cameras" ||
					(pathname.startsWith("/cameras/") && !pathname.startsWith("/cameras/new")),
			},
			{ href: "/cameras/new", label: "Add" },
		],
	},
	{
		id: "events",
		match: (pathname) => pathname === "/events" || pathname.startsWith("/events/"),
		title: "Events",
		subRoutes: [],
	},
	{
		id: "settings",
		match: (pathname) => pathname === "/settings" || pathname.startsWith("/settings/"),
		title: "Settings",
		subRoutes: [
			{ href: "/settings", label: "General", exact: true },
			{ href: "/settings/users", label: "Users", adminOnly: true },
		],
	},
];

export function resolveMainSection(pathname: string): MainSection | null {
	return sections.find((section) => section.match(pathname)) ?? null;
}

export function isSubRouteActive(pathname: string, route: SubRoute): boolean {
	if (route.isActive) {
		return route.isActive(pathname);
	}
	if (route.exact) {
		return pathname === route.href;
	}
	return pathname === route.href || pathname.startsWith(`${route.href}/`);
}
