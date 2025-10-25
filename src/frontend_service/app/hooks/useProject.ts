import { useState } from "react";
import { useNavigate } from "react-router";
import { getProjectById } from "~/services/project_service";

export default function useProject(projectId: string) {
    const [isLoading, setIsLoading] = useState(false);
    const [project, setProject] = useState<Project | null>(null);

    const loadProject = async () => {
        setIsLoading(true);

        const getProjectByIdResult = await getProjectById(projectId);

        if (getProjectByIdResult.isSuccess()) {
            setProject(getProjectByIdResult.value!);
        }

        setIsLoading(false);
    }

    return {
        isLoading,
        project,
        loadProject,
    };
}