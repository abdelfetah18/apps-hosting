import { useState } from "react";
import { getAppLogs } from "~/services/app_service";

export default function useAppLogs(projectId: string, appId: string) {
    const [logs, setLogs] = useState("");

    const loadAppLogs = async () => {
        const getAppLogsResult = await getAppLogs(projectId, appId)
        if (getAppLogsResult.isFailure()) {
            return;
        }

        setLogs(getAppLogsResult.value!);
    }

    return {
        logs,
        loadAppLogs,
    };
}