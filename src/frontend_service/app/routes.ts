import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
    route("/sign_up", "routes/sign_up.tsx"),
    route("/sign_in", "routes/sign_in.tsx"),
    layout("components/wrappers/auth_wrapper.tsx", [
        layout("components/wrappers/header_wrapper.tsx", [
            index("routes/home.tsx"),
            route("/user/settings", "routes/user_settings.tsx"),

            route("/projects/create", "routes/create_project.tsx"),
            route("/projects/:project_id/apps/create", "routes/create_app.tsx"),

            route("/callbacks/github", "routes/github_callback.tsx"),

            layout("components/wrappers/dashboard_wrapper.tsx", [
                route("/projects/:project_id", "routes/project.tsx"),
                route("/projects/:project_id/settings", "routes/project_settings.tsx"),

                route("/projects/:project_id/apps/:app_id", "routes/app.tsx"),
                route("/projects/:project_id/apps/:app_id/logs", "routes/logs.tsx"),
                route("/projects/:project_id/apps/:app_id/environment", "routes/environment.tsx"),
                route("/projects/:project_id/apps/:app_id/settings", "routes/app_settings.tsx"),
            ]),
        ]),
    ]),
] satisfies RouteConfig;
