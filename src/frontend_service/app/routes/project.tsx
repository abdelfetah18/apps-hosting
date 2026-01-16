import moment from "moment";
import type { Route } from "./+types/project";
import { Link, useOutletContext } from "react-router";
import Iconify from "~/components/Iconify";
import { RUNTIMES_LOGOS } from "~/consts";
import { getProjectById } from "~/services/project_service";
import { listApps } from "~/services/app_service";
import { useState } from "react";
import { textContainsQuery } from "~/helpers/utils";

export async function clientLoader({ params }: Route.LoaderArgs) {
    const projectId = params.project_id;

    const [projectResult, appsResult] = await Promise.all([
        getProjectById(projectId),
        listApps(projectId),
    ]);

    if (projectResult.isFailure() || appsResult.isFailure()) {
        throw new Response("Failed to load project or apps", { status: 500 });
    }

    return {
        project: projectResult.value!,
        apps: appsResult.value!,
    };
}

export default function Project({ params, loaderData }: Route.ComponentProps) {
    const [query, setQuery] = useState("");

    const project = loaderData.project;
    const apps = loaderData.apps.filter(app => textContainsQuery(app.name, query));

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                {
                    project && (
                        <div className="flex flex-col gap-4">
                            <div className="w-full flex items-center gap-4">
                                <Link to={`/projects/${params.project_id}/apps/create`} className="w-fit bg-purple-700 px-8 py-1 rounded-lg text-sm text-white font-medium">New App</Link>
                                <div className="flex items-center gap-2 text-sm border border-gray-400 rounded-lg px-4 py-1">
                                    <Iconify icon="tabler:search" />
                                    <input
                                        type="text"
                                        placeholder="Search for an app"
                                        className="outline-none border-none"
                                        value={query}
                                        onChange={(ev) => setQuery(ev.target.value)}
                                    />
                                </div>
                            </div>
                            <div className="w-full flex flex-col gap-2">
                                {
                                    apps.map((app, index) => {
                                        return (
                                            <div key={index} className="w-full flex flex-col border border-gray-200 rounded-lg px-8 py-4 gap-1">
                                                <div className="flex items-center gap-2">
                                                    <Link to={`/projects/${params.project_id}/apps/${app.id}`} className="text-balance font-semibold hover:underline">{app.name}</Link>
                                                    <div className="flex items-center gap-1">
                                                        <Iconify icon={RUNTIMES_LOGOS[app.runtime]} size={16} />
                                                        <div className="text-black text-xs">{app.runtime}</div>
                                                    </div>
                                                    <div className="flex-grow"></div>
                                                    <div className="text-gray-600 text-xs">{moment(app.created_at).format("dddd DD, MMMM YYYY HH:MM:ss")}</div>
                                                </div>
                                            </div>
                                        )
                                    })
                                }
                            </div>
                        </div >
                    )
                }
            </div >
        </div >
    );
}
