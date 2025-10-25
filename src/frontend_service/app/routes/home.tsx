import type { Route } from "./+types/home";
import { Link } from "react-router";
import { useEffect } from "react";
import moment from "moment";
import useProjects from "~/hooks/useProjects";
import Iconify from "~/components/Iconify";
import { listProjects } from "~/services/project_service";

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Home" },
    { name: "description", content: "Welcome to Apps Hosting" },
  ];
}

export async function clientLoader() {
  const listProjectsResult = await listProjects();
  if (listProjectsResult.isFailure()) {
    throw new Response("Failed to load project or apps", { status: 500 });
  }

  return {
    projects: listProjectsResult.value!,
  };
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const projects = loaderData.projects;

  return (
    <div className="w-5/6 flex flex-col gap-4 p-8">
      <div className="text-black text-xl font-semibold">Your Projects</div>
      <div className="w-full flex items-center gap-4">
        <Link to={"/projects/create"} className="w-fit bg-purple-700 px-8 py-1 rounded-lg text-sm text-white font-medium">New Project</Link>
        <div className="flex items-center gap-2 text-sm border border-gray-400 rounded-lg px-4 py-1">
          <Iconify icon="tabler:search" />
          <input type="text" placeholder="Search for an app" className="outline-none border-none" />
        </div>
      </div>
      <div className="w-full grid grid-cols-3 gap-2">
        {
          projects.map((project, index) => {
            return (
              <div key={index} className="flex flex-col border border-gray-300 rounded-lg px-4 py-4 gap-1">
                <div className="flex items-center gap-2">
                  <Link to={`/projects/${project.id}`} className="text-balance font-semibold hover:underline">{project.name}</Link>
                  <div className="flex-grow"></div>
                  <div className="text-gray-600 text-xs">{moment(project.created_at).format("dddd DD, MMMM YYYY HH:MM:ss")}</div>
                </div>
                <div className="text-xs text-gray-500">5 Apps</div>
              </div>
            )
          })
        }
      </div>
    </div>
  );
}
