import type { Route } from "./+types/home";
import { Link } from "react-router";
import { useEffect } from "react";
import useApps from "~/hooks/useApps";
import Iconify from "~/components/Iconify";
import { RUNTIMES_LOGOS } from "~/consts";
import moment from "moment";

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Apps" },
    { name: "description", content: "Welcome to Apps Hosting" },
  ];
}

export default function Apps({ params}: Route.ComponentProps) {
  const { isLoading, apps, loadApps } = useApps();

  useEffect(() => {
    loadApps();
  }, []);

  return (
    <div className="w-full flex flex-col">
      <div className="text-black text-xl font-semibold">My Apps</div>
      <div className="w-full flex flex-col gap-2">
        {
          isLoading && (<div>Loading...</div>)
        }
        {
          apps.map((app, index) => {
            return (
              <div key={index} className="flex flex-col border-b last:border-b-0 border-gray-200 py-4 gap-1">
                <div className="flex items-center gap-2">
                  <Link to={`/apps/${app.id}`} className="text-balance font-semibold hover:underline">{app.name}</Link>
                  <div className="flex items-center gap-1">
                    <Iconify icon={RUNTIMES_LOGOS[app.runtime]} size={16} />
                    <div className="text-black text-xs">{app.runtime}</div>
                  </div>
                  <div className="flex-grow"></div>
                  <div className="text-gray-600 text-xs">{moment(app.created_at).format("dddd DD, MMMM YYYY HH:MM:ss")}</div>
                </div>

                {/* <div className={`flex items-center gap-1 ${APP_STATUS_COLORS[app.status]} text-xs font-semibold`}>
                  <Iconify icon={APP_STATUS_ICONS[app.status]} size={16} />
                  <span className="capitalize">{app.status.replaceAll("_", " ")}</span>
                </div> */}
              </div>
            )
          })
        }
      </div>
    </div>
  );
}
