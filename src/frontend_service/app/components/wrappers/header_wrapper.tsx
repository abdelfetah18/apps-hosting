import { Link, Outlet, useLocation, useOutletContext } from "react-router";
import SideBar from "~/components/sidebar";
import type { Route } from "./+types/dashboard_wrapper";
import { getSidebarItems } from "~/helpers/config";
import { DEFAULT_PROFILE_PICTURE } from "~/consts";
import Iconify from "../Iconify";
import { useState } from "react";

export default function HeaderWrapper({ params }: Route.ComponentProps) {
    const location = useLocation();
    const outletData: UserSession = useOutletContext();
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className="w-full h-screen overflow-auto flex flex-col items-center">
            <div className="w-full flex items-center justify-between px-8 py-2 border-b border-gray-300">
                <div className="flex items-center gap-4">
                    <Link to="/">Logo</Link>
                    <div className="flex items-center gap-1">
                        <Link to={`/projects/${params.project_id}`} className="text-xs text-gray-700 underline">{params.project_id}</Link>
                        <div>{"/"}</div>
                        <div className="text-xs text-gray-700 underline">{params.app_id}</div>
                    </div>
                </div>
                <div className="w-8 h-8 relative">
                    <img src={DEFAULT_PROFILE_PICTURE} className="w-full h-full object-cover rounded-full cursor-pointer" onClick={() => setIsOpen(state => !state)} />
                    {
                        isOpen && (
                            <div className="absolute top-full my-2 right-0 border border-gray-300 shadow-2xl bg-white flex flex-col px-8 py-4">
                                <div className="w-full flex items-center gap-2">
                                    <div className="h-10 w-10 rounded-full bg-gray-300"></div>
                                    <div>
                                        <div className="">{outletData.user.username}</div>
                                        <div className="text-xs">{outletData.user.email}</div>
                                    </div>
                                </div>
                                <div className="w-full flex items-center gap-2 text-sm border-b border-gray-300 py-4">
                                    <Iconify icon="quill:cog-alt" size={20} />
                                    <div className="font-medium">Account setting</div>
                                </div>
                                <div className="w-full flex items-center gap-2 text-sm py-4">
                                    <Iconify icon="material-symbols:logout-rounded" size={20} />
                                    <div className="font-medium">Sign Out</div>
                                </div>
                            </div>
                        )
                    }
                </div>
            </div>
            <div className="w-full h-full flex flex-col items-center overflow-auto">
                <Outlet context={outletData} />
            </div>
        </div>
    );
}
