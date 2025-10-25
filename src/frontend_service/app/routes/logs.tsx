import useAppLogs from "~/hooks/useAppLogs";
import type { Route } from "./+types/logs";
import { useEffect, useRef } from "react";
import AnsiLogs from "~/components/AnsiLogs";

export default function Logs({ params }: Route.ComponentProps) {
    const { logs, loadAppLogs } = useAppLogs(params.project_id, params.app_id);
    const logsWrapperRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        loadAppLogs();
    }, [])

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full h-full flex flex-col gap-2">
                <div className="h-full flex flex-col gap-4">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">Logs</div>
                    </div>

                    <div ref={logsWrapperRef} className="w-full h-96 overflow-auto flex flex-col border border-gray-400 p-2 text-xs whitespace-pre-line">
                        <AnsiLogs logs={logs} />
                    </div>
                </div>
            </div>
        </div>
    );
}
