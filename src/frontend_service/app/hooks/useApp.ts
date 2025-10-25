import { useState } from "react";
import { useNavigate } from "react-router";
import { deleteAppById, getAppById, reDeployAppById, listDeployments, listBuilds } from "~/services/app_service";

export default function useApp(projectId: string, appId: string) {
    const [isLoading, setIsLoading] = useState(false);
    const [app, setApp] = useState<App | null>(null);
    const [deployments, setDeployments] = useState<Deployment[]>([]);
    const [builds, setBuilds] = useState<Build[]>([]);
    const navigate = useNavigate();

    const loadApp = async () => {
        setIsLoading(true);

        const [getAppByIdResult, listDeploymentsResult, listBuildsResult] = await Promise.all([
            getAppById(projectId, appId),
            listDeployments(projectId, appId),
            listBuilds(projectId, appId)
        ]);

        if (getAppByIdResult.isSuccess()) {
            setApp(getAppByIdResult.value!);
        }

        if (listDeploymentsResult.isSuccess()) {
            setDeployments(listDeploymentsResult.value!);
        }

        if (listBuildsResult.isSuccess()) {
            setBuilds(listBuildsResult.value!);
        }

        setIsLoading(false);
    }

    const reDeployApp = async () => {
        const result = await reDeployAppById(projectId, appId);
        if (result.isFailure()) {
            alert(result.error);
        } else {
            alert(result.value);
        }
    }


    const deletedAppHandler = async () => {
        const deleteAppByIdResult = await deleteAppById(projectId, appId)
        if (deleteAppByIdResult.isSuccess()) {
            navigate("/");
        } else {
            console.error(deleteAppByIdResult.error);
        }
    }

    return {
        isLoading,
        app,
        loadApp,
        reDeployApp,
        deployments,
        builds,
        deletedAppHandler,
    };
}