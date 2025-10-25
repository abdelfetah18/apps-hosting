import { useState } from "react";
import { listProjects } from "~/services/project_service";

export default function useProjects() {
    const [isLoading, setIsLoading] = useState(false);
    const [projects, setProjects] = useState<Project[]>([]);

    const loadProjects = async () => {
        setIsLoading(true);
        const result = await listProjects();
        if (result.isFailure()) {
            return;
        }

        setProjects(result.value!);
        setIsLoading(false);
    }

    return {
        isLoading,
        projects,
        loadProjects,
    };
}