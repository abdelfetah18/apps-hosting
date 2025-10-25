import { Outlet, useLocation, useOutletContext } from "react-router";
import SideBar from "~/components/sidebar";
import type { Route } from "./+types/dashboard_wrapper";
import { getSidebarItems } from "~/helpers/config";

export default function DashboardWrapper({ params }: Route.ComponentProps) {
    const location = useLocation();
    const outletData = useOutletContext();

    const pathname = location.pathname;

    return (
        <div className="w-full flex-grow overflow-auto flex">
            <div className="w-1/6 h-full border-r border-gray-200">
                <SideBar
                    goBackTo={pathname != "/" ? {
                        path: "/",
                        title: "Dashboard"
                    } : undefined}
                    items={getSidebarItems(pathname, params)}
                />
            </div>
            <div className="w-5/6 flex flex-col gap-8 overflow-auto">
                <Outlet context={outletData} />
            </div>
        </div>
    );
}
