import { Link, useOutletContext } from "react-router";
import Iconify from "./Iconify";
import { DEFAULT_PROFILE_PICTURE } from "~/consts";

interface Item {
    name: string;
    icon: string;
    path: string;
}

interface SideBarProps {
    items: Item[];
    goBackTo?: {
        title: string;
        path: string;
    };
}

export default function SideBar({ items }: SideBarProps) {
    const userSession = useOutletContext() as UserSession;

    return (
        <div className="w-full flex flex-col items-center gap-16 p-4">
            <div className="w-full flex flex-col">
                {
                    items.map((item, index) => {
                        const isSelected = location.pathname == item.path;

                        return (
                            <Link key={index} to={item.path} className={`w-full flex items-center px-4 py-2 gap-2 text-sm font-semibold hover:text-purple-600 hover:underline cursor-pointer duration-300 rounded ${isSelected ? "text-purple-700 bg-purple-700/10" : "text-black"}`}>
                                <Iconify icon={item.icon} size={20} /> {item.name}
                            </Link>
                        )
                    })
                }
            </div>
        </div>
    )
}