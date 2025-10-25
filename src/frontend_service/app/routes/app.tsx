import type { Route } from "./+types/app";
import { Link } from "react-router";
import useApp from "~/hooks/useApp";
import { useEffect } from "react";
import Iconify from "~/components/Iconify";
import { BUILD_STATUS_COLORS, BUILD_STATUS_ICONS, DEPLOYMENT_STATUS_COLORS, DEPLOYMENT_STATUS_ICONS, RUNTIMES_LOGOS } from "~/consts";
import moment from "moment";
import { toActivityList } from "../helpers/utils";

export default function App({ params }: Route.ComponentProps) {
    const { isLoading, app, deployments, builds, loadApp, reDeployApp, deletedAppHandler } = useApp(params.project_id, params.app_id as string);

    const reDeployHandler = () => {
        reDeployApp();
    }

    useEffect(() => {
        loadApp();
    }, []);

    const copyToClipboardHandler = () => {
        navigator.clipboard.writeText(`https://${app?.domain_name}`);
    }

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                {
                    isLoading && (<div>Loading...</div>)
                }
                {
                    app && (
                        <div className="flex flex-col gap-4">
                            <div className="flex items-start justify-between">
                                <div className="flex items-center gap-2">
                                    <div className="text-xl text-balance font-semibold">{app.name}</div>
                                    <div className="flex items-center gap-1">
                                        <Iconify icon={RUNTIMES_LOGOS[app.runtime]} size={16} />
                                        <div className="text-black text-xs">{app.runtime}</div>
                                    </div>
                                </div>
                                <div className="flex items-center gap-2">
                                    <div onClick={deletedAppHandler} className="w-fit flex items-center gap-2 text-white bg-black px-8 py-2 cursor-pointer active:scale-105 select-none duration-300 hover:bg-red-500">
                                        <Iconify icon="material-symbols:delete-outline" />
                                        <div>Delete</div>
                                    </div>
                                    <div onClick={reDeployHandler} className="w-fit text-white bg-black px-8 py-2 cursor-pointer active:scale-105 select-none duration-300 hover:bg-red-500">ReDeploy</div>
                                </div>
                            </div>

                            <div className="flex items-center gap-1">
                                <Iconify icon="mdi:github" size={16} />
                                <Link to={app.repo_url} target="_blank" className="text-gray-800 text-xs underline">{app.repo_url}</Link>
                            </div>
                            <div className="flex items-center gap-1 text-purple-700">
                                <Link to={`https://${app.domain_name}`} target="_blank" className="text-xs underline">{`https://${app.domain_name}`}</Link>
                                <div className="cursor-pointer active:scale-110 duration-300 select-none" onClick={copyToClipboardHandler}>
                                    <Iconify icon="lucide:copy" size={16} />
                                </div>
                            </div>

                            {/* <div className="text-black text-xs font-semibold">Building Status: <span className="text-green-700">{app.build.status}</span></div>
                            <div className="text-black text-xs font-semibold">Deployment Status: <span className="text-green-700">{app.build.deployment.status}</span></div>
                            <div className="text-black text-xs font-semibold">Build Command: npm run {app.build_cmd}</div>
                            <div className="text-black text-xs font-semibold">Start Command: npm run {app.start_cmd}</div>
                            <div className="w-full bg-black text-white text-xs p-2 whitespace-pre">Logs...</div>
                             */}

                            <div className="w-full flex flex-col border border-gray-400">
                                {
                                    toActivityList(builds, deployments).map((activity, index) => {
                                        const isDeployment = activity.type == "deployment";

                                        return (
                                            <div key={index} className="flex items-center border-b last:border-b-0 border-gray-400 p-4">
                                                {
                                                    isDeployment ? (
                                                        <div className={`flex items-center gap-1 ${DEPLOYMENT_STATUS_COLORS[activity.status]} px-4`}>
                                                            <Iconify icon={DEPLOYMENT_STATUS_ICONS[activity.status]} />
                                                        </div>
                                                    ) : (
                                                        <div className={`flex items-center gap-1 ${BUILD_STATUS_COLORS[activity.status]} px-4`}>
                                                            <Iconify icon={BUILD_STATUS_ICONS[activity.status]} />
                                                        </div>
                                                    )
                                                }
                                                {
                                                    isDeployment ? (
                                                        <div className="text-sm text-gray-900">Deploy live for <span className="underline cursor-pointer hover:text-purple-800">{activity.build_id}</span></div>
                                                    ) : (
                                                        <div className="text-sm text-gray-900">Build {activity.status} <a className="underline cursor-pointer hover:text-purple-800">{activity.commit_hash}</a></div>
                                                    )
                                                }
                                                <div className="flex-grow"></div>
                                                <div className="text-xs text-gray-600">{moment(activity.created_at).fromNow()}</div>
                                            </div>
                                        );
                                    })
                                }
                            </div>
                        </div>
                    )
                }
            </div>
        </div>
    );
}
