export function getSidebarItems(pathname: string, params: { [key: string]: string | undefined; }) {
    if (pathname === "/") {
        return [
            { icon: "charm:apps-plus", name: "Create Project", path: "/projects/create" },
            { icon: "charm:apps", name: "My Projects", path: "/" },
        ];
    }

    const projectId = params["project_id"];
    const appId = params["app_id"];

    if (appId && pathname.startsWith(`/projects/${projectId}/apps/${appId}`)) {
        return [
            { icon: "material-symbols:deployed-code-outline", name: "Deployments", path: `/projects/${projectId}/apps/${appId}` },
            { icon: "icon-park-outline:log", name: "Logs", path: `/projects/${projectId}/apps/${appId}/logs` },
            { icon: "uil:file-lock-alt", name: "Environment", path: `/projects/${projectId}/apps/${appId}/environment` },
            { icon: "quill:cog-alt", name: "Settings", path: `/projects/${projectId}/apps/${appId}/settings` },
        ];
    }

    if (projectId && pathname.startsWith(`/projects/${projectId}`)) {
        return [
            { icon: "charm:apps", name: "My Apps", path: `/projects/${projectId}` },
            { icon: "quill:cog-alt", name: "Settings", path: `/projects/${projectId}/settings` },
        ];
    }

    return [
        { icon: "charm:apps-plus", name: "Create Project", path: "/projects/create" },
        { icon: "charm:apps", name: "My Projects", path: "/" },
    ];
}
