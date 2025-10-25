import { createApp, createEnvironmentVariables } from "~/services/app_service";
import type { Route } from "./+types/create_app";
import { redirect, useFetcher } from "react-router";
import { RUNTIMES } from "~/consts";
import { useContext, useEffect, useState } from "react";
import Iconify from "~/components/Iconify";
import Spinner from "~/components/spinner";
import { createProject } from "~/services/project_service";
import ToastContext from "~/contexts/ToastContext";

export async function clientAction({ request }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const createProjectForm: CreateProjectForm = {
        name: formData.get("name")?.toString() || "",
    };

    const errors: Record<string, string> = {};

    if (createProjectForm.name.length === 0) {
        errors.name = "Project name is required.";
    }

    if (Object.keys(errors).length > 0) {
        return errors;
    }

    const result = await createProject(createProjectForm);
    if (result.isFailure()) {
        errors.api = result.error!;
        return errors;
    }

    return redirect(`/projects/${result.value?.id}`);
}

export default function CreateApp({ }: Route.ComponentProps) {
    const toastManager = useContext(ToastContext);
    const fetcher = useFetcher();

    useEffect(() => {
        if (fetcher.data?.api != undefined) {
            toastManager.alertError(fetcher.data.api);
        }
    }, [fetcher.data]);

    return (
        <div className="w-2/3 flex flex-col gap-8 p-8 border border-gray-300 rounded-lg my-8">
            <div className="text-black text-xl font-semibold">Create App</div>
            <fetcher.Form method="POST" className="w-full flex flex-col gap-8">
                <div className="w-full flex flex-col gap-4">

                    {/* Name */}
                    <div className="flex">
                        <div className="w-1/3 flex flex-col gap-1 text-black">
                            <div>Name</div>
                            <div className="text-sm text-gray-500">A unique name for your project.</div>
                        </div>
                        <div className="h-fit flex-grow flex flex-col gap-2">
                            <input
                                name="name"
                                type="text"
                                placeholder="App name"
                                className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.name ? "border-red-300" : "border-gray-300"}`}
                            />
                            <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                        </div>
                    </div>
                </div>
                <div className="w-full flex flex-col items-end">
                    {
                        fetcher.state == "idle" ? (

                            <button
                                type="submit"
                                className="w-96 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none bg-purple-700 text-white text-center rounded-full py-2 font-semibold"
                            >
                                Create Project
                            </button>
                        ) : (
                            <button
                                type="button"
                                className="w-96 flex items-center justify-center cursor-pointer select-none py-2"
                            >
                                <Spinner />
                            </button>
                        )
                    }
                </div>

            </fetcher.Form>
        </div>
    );
}
