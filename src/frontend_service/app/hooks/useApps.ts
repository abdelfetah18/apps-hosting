import { useState } from "react";
import { listApps } from "~/services/app_service";

export default function useApps(projectId: string) {
    const [isLoading, setIsLoading] = useState(false);
    const [apps, setApps] = useState<App[]>([]);

    const loadApps = async () => {
        setIsLoading(true);
        const accessToken = window.localStorage.getItem("token") as string;

        const result = await listApps(projectId);
        if (result.isFailure()) {
            return;
        }

        setApps(result.value!);
        setIsLoading(false);
    }

    return {
        isLoading,
        apps,
        loadApps,
    };
}