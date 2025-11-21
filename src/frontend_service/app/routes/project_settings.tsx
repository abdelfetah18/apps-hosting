import { redirect, useFetcher } from "react-router";
import type { Route } from "./+types/project_settings";
import { getProjectById, updateProjectById } from "~/services/project_service";
import { useState } from "react";
import Spinner from "~/components/spinner";

export async function clientAction({ request, params }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const updateProjectForm: UpdateProjectForm = {
        name: formData.get("name")?.toString() || "",
    };

    const errors: Record<string, string> = {};

    if (updateProjectForm.name.length === 0) {
        errors.name = "Project name is required.";
    }

    if (Object.keys(errors).length > 0) {
        return errors;
    }

    const result = await updateProjectById(params.project_id, updateProjectForm);
    if (result.isFailure()) {
        errors.apiError = result.error!;
        return errors;
    }

    return {};
}

export async function clientLoader({ params }: Route.LoaderArgs) {
    const projectId = params.project_id;

    const projectResult = await getProjectById(projectId);

    if (projectResult.isFailure()) {
        throw new Response("Failed to load project", { status: 500 });
    }

    return {
        project: projectResult.value!,
    };
}

export default function ProjectSettings({ params, loaderData }: Route.ComponentProps) {
    let errorMessage = undefined;
    const fetcher = useFetcher();

    const [name, setName] = useState(loaderData.project.name);
    const isDraft = name !== loaderData.project.name;

    const resetNameInput = () => {
        setName(loaderData.project.name);
        (document.querySelector("input[name='name']") as HTMLInputElement).value = loaderData.project.name;
    }

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                <div className="flex flex-col gap-8">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">Project settings</div>
                    </div>
                    <fetcher.Form method="POST" className="w-full flex flex-col gap-4">
                        {/* Name */}
                        <div className="flex">
                            <div className="w-1/3 flex flex-col gap-1 text-black">
                                <div>Name</div>
                                <div className="text-sm text-gray-500">A unique name for your project.</div>
                            </div>
                            <div className="h-fit grow flex flex-col gap-2">
                                <input
                                    name="name"
                                    type="text"
                                    placeholder="Name"
                                    defaultValue={loaderData.project.name}
                                    onChange={(e) => setName(e.target.value)}
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.name ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                            </div>
                        </div>

                        <div className="w-full flex items-center justify-end">
                            {
                                isDraft && (
                                    <div className="flex items-center gap-2">
                                        <div onClick={resetNameInput} className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none border border-purple-700 text-purple-700 text-center rounded-full py-1 font-semibold">Cancel</div>
                                        {
                                            fetcher.state == "idle" ? (

                                                <button
                                                    type="submit"
                                                    className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none bg-purple-700 text-white text-center rounded-full py-1 font-semibold"
                                                >
                                                    Save
                                                </button>
                                            ) : (
                                                <button
                                                    type="button"
                                                    className="w-30 flex items-center justify-center cursor-pointer select-none py-2"
                                                >
                                                    <Spinner />
                                                </button>
                                            )
                                        }
                                    </div>
                                )
                            }
                        </div>
                    </fetcher.Form>
                </div>
            </div>
        </div>
    );
}
