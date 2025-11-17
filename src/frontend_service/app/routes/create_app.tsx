import { createApp, createEnvironmentVariables } from "~/services/app_service";
import type { Route } from "./+types/create_app";
import { Link, redirect, useFetcher, useOutletContext } from "react-router";
import { RUNTIMES } from "~/consts";
import { useEffect, useState } from "react";
import Iconify from "~/components/Iconify";
import Spinner from "~/components/spinner";
import { getUserGithubRepositories } from "~/services/user_service";

export async function clientAction({ request }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const keys = formData.getAll("env_key") as string[];
    const values = formData.getAll("env_value") as string[];
    const projectId = formData.get("project_id")?.toString() || "";

    const createAppForm: CreateAppForm = {
        name: formData.get("name")?.toString() || "",
        runtime: formData.get("runtime")?.toString() || "",
        repo_url: formData.get("repo_url")?.toString() || "",
        start_cmd: formData.get("start_cmd")?.toString() || "",
        build_cmd: formData.get("build_cmd")?.toString() || "",
        git_repository: {
            clone_url: formData.get("repo_url")?.toString() || "",
            is_private: (formData.get("git_repo_is_private")?.toString() || "") == "private",
            provider: (formData.get("git_repo_provider")?.toString() as GitProvider) || "github",
        }

    };

    const errors: Record<string, string> = {};

    if (createAppForm.name.length === 0) {
        errors.name = "App name is required.";
    }

    if (createAppForm.repo_url.length === 0) {
        errors.repo_url = "Repository URL is required.";
    } else {
        const isGitHubRepo = /^https:\/\/(www\.)?github\.com\/[\w.-]+\/[\w.-]+(?:\/)?$/.test(createAppForm.repo_url);

        if (!isGitHubRepo) {
            errors.repo_url = "Please provide a valid GitHub repository URL.";
        }
    }

    if (createAppForm.runtime.length === 0) {
        errors.runtime = "Please select a runtime.";
    }

    if ((createAppForm.runtime as Runtime) === "NodeJS") {
        if (createAppForm.start_cmd.length === 0) {
            errors.start_cmd = "Start command is required for NodeJS runtime.";
        }

        if (createAppForm.build_cmd.length === 0) {
            errors.build_cmd = "Build command is required for NodeJS runtime.";
        }
    }

    if (Object.keys(errors).length > 0) {
        return errors;
    }

    const result = await createApp(projectId, createAppForm);
    if (result.isFailure()) {
        errors.apiError = result.error!;
        return errors;
    }

    if (keys.length > 0) {
        const value = JSON.stringify(
            Object
                .fromEntries(
                    keys.map((key, i) => [key.trim(), values[i]?.trim()])
                        .filter(([key, value]) => key && value)));
        const envVarsResult = await createEnvironmentVariables(projectId, result.value!.id, value);
        if (envVarsResult.isFailure()) {
            errors.apiError = envVarsResult.error!;
            return errors;
        }
    }

    return redirect(`/projects/${projectId}/apps/${result.value?.id}`);
}

export async function clientLoader() {
    const getUserGithubRepositoriesResult = await getUserGithubRepositories();
    if (getUserGithubRepositoriesResult.isFailure()) {
        return {
            githubRepositories: [],
        };
    }


    return {
        githubRepositories: getUserGithubRepositoriesResult.value!,
    };
}

export default function CreateApp({ params, loaderData }: Route.ComponentProps) {
    const userSession: UserSession = useOutletContext();
    const fetcher = useFetcher();
    const [envVars, setEnvVars] = useState([{ key: "", value: "" }]);
    const githubRepositories = loaderData.githubRepositories;

    const handleAddEnvVar = () => {
        setEnvVars([...envVars, { key: "", value: "" }]);
    };

    const handleEnvVarChange = (index: number, field: "key" | "value", value: string) => {
        const updated = [...envVars];
        updated[index][field] = value;
        setEnvVars(updated);
    };

    const handleRemoveEnvVar = (index: number) => {
        setEnvVars(envVars.filter((_, i) => i !== index));
    };

    const configureGithub = () => {
        window.open(new URL("https://github.com/apps/apps-hosting/installations/new"));
    }

    useEffect(() => {
        window.onmessage = (event: MessageEvent) => {
            if (event.data == "github-callback-done") {
                window.location.reload();
            }
        }
    }, []);

    return (
        <div className="w-2/3 flex flex-col gap-8 p-8 border border-gray-300 rounded-lg my-8">
            <div className="text-black text-xl font-semibold">Create App</div>
            <fetcher.Form method="POST" className="w-full flex flex-col gap-8">
                <div className="w-full flex flex-col gap-4">
                    { /* Project Id */}
                    <input
                        name="project_id"
                        value={params.project_id}
                        type="hidden"
                    />
                    {/* Name */}
                    <div className="flex">
                        <div className="w-1/3 flex flex-col gap-1 text-black">
                            <div>Name</div>
                            <div className="text-sm text-gray-500">A unique name for your web service.</div>
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

                    {/* Runtime */}
                    <div className="flex">
                        <div className="w-1/3 text-black">Runtime</div>
                        <div className="h-fit flex-grow flex flex-col gap-2">
                            <select
                                name="runtime"
                                className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.runtime ? "border-red-300" : "border-gray-300"}`}
                            >
                                <option disabled selected>Runtime</option>
                                {RUNTIMES.map((runtime, index) => (
                                    <option key={index} value={runtime}>{runtime}</option>
                                ))}
                            </select>
                            <div className="text-red-600 text-xs">{fetcher.data?.runtime}</div>
                        </div>
                    </div>

                    {/* Repo URL */}
                    <div className="flex">
                        <div className="w-1/3 text-black">Repo URL</div>
                        <div className="h-fit flex-grow flex flex-col gap-2">
                            <div className="w-full flex flex-col gap-2">
                                <div className="w-full flex items-center gap-2">
                                    <input name="git_repo_is_private" hidden />
                                    <input name="git_repo_provider" value={"github"} hidden />
                                    <input
                                        name="repo_url"
                                        type="text"
                                        placeholder="Repo URL"
                                        className={`flex-grow border rounded-lg px-4 py-2 text-sm ${fetcher.data?.repo_url ? "border-red-300" : "border-gray-300"}`}
                                    />
                                    {
                                        userSession.user.github_app_installed ? (
                                            <div onClick={configureGithub} className="bg-black text-white text-sm font-semibold px-4 py-2 rounded-lg cursor-pointer active:scale-105 duration-300 select-none">Configure</div>
                                        ) : (
                                            <div onClick={configureGithub} className="flex items-center justify-center gap-2 bg-black text-white text-sm font-semibold px-4 py-2 rounded-lg cursor-pointer active:scale-105 duration-300 select-none">
                                                <Iconify icon="mdi:github" size={16} />
                                                <div>Connect to Github</div>
                                            </div>
                                        )
                                    }
                                </div>
                                {
                                    userSession.user.github_app_installed && (
                                        <div className="w-full h-60 overflow-auto top-full left-0 flex flex-col border border-gray-300 rounded-lg bg-white">
                                            {
                                                githubRepositories.map((repo, index) => {
                                                    const selectRepo = () => {
                                                        (document.querySelector("[name=repo_url]") as HTMLInputElement).value = repo.clone_url;
                                                        (document.querySelector("[name=git_repo_is_private]") as HTMLInputElement).value = repo.visibility;
                                                    }

                                                    return (
                                                        <div onClick={selectRepo} key={index} className="px-4 py-2 text-sm hover:bg-gray-100 cursor-pointer flex items-center gap-2">
                                                            <Iconify icon="mdi:github" size={16} />
                                                            <div className="text-sm">{repo.name}</div>
                                                            <div className="text-sm">{repo.visibility}</div>
                                                        </div>
                                                    )
                                                })
                                            }
                                        </div>
                                    )
                                }
                            </div>
                            <div className="text-red-600 text-xs">{fetcher.data?.repo_url}</div>
                        </div>
                    </div>

                    {/* Build Command */}
                    <div className="flex">
                        <div className="w-1/3 text-black">Build Command</div>
                        <div className="h-fit flex-grow flex flex-col gap-2">
                            <input
                                name="build_cmd"
                                defaultValue="build"
                                type="text"
                                placeholder="Build Command"
                                className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.build_cmd ? "border-red-300" : "border-gray-300"}`}
                            />
                            <div className="text-red-600 text-xs">{fetcher.data?.build_cmd}</div>
                        </div>
                    </div>

                    {/* Start Command */}
                    <div className="flex">
                        <div className="w-1/3 text-black">Start Command</div>
                        <div className="h-fit flex-grow flex flex-col gap-2">
                            <input
                                name="start_cmd"
                                defaultValue="start"
                                type="text"
                                placeholder="Start Command"
                                className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.start_cmd ? "border-red-300" : "border-gray-300"}`}
                            />
                            <div className="text-red-600 text-xs">{fetcher.data?.start_cmd}</div>
                        </div>
                    </div>

                    {/* Environment Variables */}
                    <div className="flex">
                        <div className="w-1/3 text-black">Environment Variables</div>
                        <div className="w-2/3 flex flex-col gap-2">
                            {envVars.map((env, index) => (
                                <div key={index} className="flex gap-2 items-center">
                                    <input
                                        name="env_key"
                                        placeholder="Key"
                                        value={env.key}
                                        onChange={(e) => handleEnvVarChange(index, "key", e.target.value)}
                                        className="flex-1 border border-gray-300 rounded-lg px-4 py-2 text-sm"
                                    />
                                    <input
                                        name="env_value"
                                        placeholder="Value"
                                        value={env.value}
                                        onChange={(e) => handleEnvVarChange(index, "value", e.target.value)}
                                        className="flex-1 border border-gray-300 rounded-lg px-4 py-2 text-sm"
                                    />
                                    {envVars.length > 1 && (
                                        <button
                                            type="button"
                                            onClick={() => handleRemoveEnvVar(index)}
                                            className="text-red-500 cursor-pointer"
                                        >
                                            <Iconify icon="material-symbols:delete-outline" size={16} />
                                        </button>
                                    )}
                                </div>
                            ))}
                            <button
                                type="button"
                                onClick={handleAddEnvVar}
                                className="flex items-center gap-1 text-sm text-blue-600 hover:underline w-fit mt-1 cursor-pointer"
                            >
                                <Iconify icon="material-symbols:add-rounded" /> Add Variable
                            </button>
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
                                Deploy App
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
