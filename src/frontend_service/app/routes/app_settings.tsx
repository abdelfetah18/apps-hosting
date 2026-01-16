import { useFetcher, useNavigate } from "react-router";
import type { Route } from "./+types/app_settings";
import { useState } from "react";
import { deleteAppById, getAppById, updateAppById } from "~/services/app_service";
import Spinner from "~/components/spinner";
import Iconify from "~/components/Iconify";


export async function clientAction({ request, params }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const updateAppForm: UpdateAppForm = {
        name: formData.get("name")?.toString() || "",
        build_cmd: formData.get("build_cmd")?.toString() || "",
        start_cmd: formData.get("start_cmd")?.toString() || "",
    };

    const errors: Record<string, string> = {};

    if (updateAppForm.name.length === 0) {
        errors.name = "App name can not be empty.";
    }

    if (updateAppForm.build_cmd.length === 0) {
        errors.build_cmd = "App build command can not be empty.";
    }

    if (updateAppForm.start_cmd.length === 0) {
        errors.start_cmd = "App start command can not be empty.";
    }

    if (Object.keys(errors).length > 0) {
        return errors;
    }

    const result = await updateAppById(params.project_id, params.app_id, updateAppForm);
    if (result.isFailure()) {
        errors.apiError = result.error!;
        return errors;
    }

    return {};
}

export async function clientLoader({ params }: Route.LoaderArgs) {
    const projectId = params.project_id;
    const appId = params.app_id;

    const appResult = await getAppById(projectId, appId);

    if (appResult.isFailure()) {
        throw new Response("Failed to load app", { status: 500 });
    }

    return {
        app: appResult.value!,
    };
}

export default function AppSettings({ params, loaderData }: Route.ComponentProps) {
    const navigate = useNavigate();
    const [isDeleteAppLoading, setIsDeleteAppLoading] = useState(false);
    const fetcher = useFetcher();

    const [updateAppForm, setUpdateAppForm] = useState<UpdateAppForm>({
        name: loaderData.app.name,
        build_cmd: loaderData.app.build_cmd,
        start_cmd: loaderData.app.start_cmd,
    });

    const isDraft = (updateAppForm.name !== loaderData.app.name) ||
        (updateAppForm.build_cmd !== loaderData.app.build_cmd) ||
        (updateAppForm.start_cmd !== loaderData.app.start_cmd);

    const resetFormInputs = () => {
        setUpdateAppForm({
            name: loaderData.app.name,
            build_cmd: loaderData.app.build_cmd,
            start_cmd: loaderData.app.start_cmd,
        });
    }

    const deleteAppHandler = async () => {
        setIsDeleteAppLoading(true);
        const result = await deleteAppById(params.project_id, params.app_id);
        if (result.isSuccess()) {
            navigate("/");
        } else {
            console.error(result.error);
        }
        setIsDeleteAppLoading(false);
    }

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                <div className="flex flex-col gap-8">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">Settings</div>
                    </div>
                    <fetcher.Form method="POST">
                        <div className="w-full flex flex-col gap-4">
                            <div className="flex items-start justify-between">
                                <div className="text-lg text-balance font-semibold hover:underline">General</div>
                            </div>
                            <div className="flex">
                                <div className="w-1/3 flex flex-col gap-1 text-black">
                                    <div>Name</div>
                                    <div className="text-sm text-gray-500">A unique name for your app.</div>
                                </div>
                                <div className="h-fit grow flex flex-col gap-2">
                                    <input
                                        name="name"
                                        type="text"
                                        placeholder="Name"
                                        value={updateAppForm.name}
                                        onChange={(e) => setUpdateAppForm(state => ({ ...state, name: e.target.value }))}
                                        className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.name ? "border-red-300" : "border-gray-300"}`}
                                    />
                                    <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                                </div>
                            </div>
                        </div>
                        <div className="w-full flex flex-col gap-4">
                            <div className="flex items-start justify-between">
                                <div className="text-lg text-balance font-semibold hover:underline">Build & Deploy</div>
                            </div>
                            <div className="flex">
                                <div className="w-1/3 flex flex-col gap-1 text-black">
                                    <div>Build Command</div>
                                    <div className="text-sm text-gray-500">We runs this command to build your app before each deploy.</div>
                                </div>
                                <div className="h-fit grow flex flex-col gap-2">
                                    <input
                                        name="build_cmd"
                                        type="text"
                                        placeholder="Build Command"
                                        value={updateAppForm.build_cmd}
                                        onChange={(e) => setUpdateAppForm(state => ({ ...state, build_cmd: e.target.value }))}
                                        className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.build_cmd ? "border-red-300" : "border-gray-300"}`}
                                    />
                                    <div className="text-red-600 text-xs">{fetcher.data?.build_cmd}</div>
                                </div>
                            </div>
                            <div className="flex">
                                <div className="w-1/3 flex flex-col gap-1 text-black">
                                    <div>Start Command</div>
                                    <div className="text-sm text-gray-500">We runs this command to start your app with each deploy.</div>
                                </div>
                                <div className="h-fit grow flex flex-col gap-2">
                                    <input
                                        name="start_cmd"
                                        type="text"
                                        placeholder="Start Command"
                                        value={updateAppForm.start_cmd}
                                        onChange={(e) => setUpdateAppForm(state => ({ ...state, start_cmd: e.target.value }))}
                                        className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.start_cmd ? "border-red-300" : "border-gray-300"}`}
                                    />
                                    <div className="text-red-600 text-xs">{fetcher.data?.start_cmd}</div>
                                </div>
                            </div>
                        </div>
                        <div className="w-full flex items-center justify-end">
                            {
                                isDraft && (
                                    <div className="flex items-center gap-2">
                                        <div onClick={resetFormInputs} className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none border border-purple-700 text-purple-700 text-center rounded-lg py-1 font-semibold">Cancel</div>
                                        {
                                            fetcher.state == "idle" ? (

                                                <button
                                                    type="submit"
                                                    className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none bg-purple-700 text-white text-center rounded-lg py-1 font-semibold"
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
                    <div className="flex items-start justify-between">
                        {
                            isDeleteAppLoading ? (
                                <button
                                    type="button"
                                    className="w-48 flex items-center justify-center cursor-pointer select-none py-2"
                                >
                                    <Spinner />
                                </button>
                            ) : (
                                <div onClick={deleteAppHandler} className="w-48 flex items-center justify-center gap-2 bg-red-500 py-2 px-4 rounded-lg text-white cursor-pointer select-none">
                                    <Iconify icon="material-symbols:delete-outline" size={20} />
                                    <div>Delete App</div>
                                </div>
                            )
                        }
                    </div>
                </div>
            </div>
        </div>
    );
}
