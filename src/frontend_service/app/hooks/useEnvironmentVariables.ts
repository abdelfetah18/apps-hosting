import { useState } from "react";
import { getEnvironmentVariables, createEnvironmentVariables, updateEnvironmentVariables } from "~/services/app_service";

export default function useEnvironmentVariables(appID: string) {
    const [isLoading, setIsLoading] = useState(false);
    const [envVars, setEnvVars] = useState<{ key: string; value: string; }[]>([]);

    const loadEnvironmentVariables = async () => {
        setIsLoading(true);

        const getEnvironmentVariablesResult = await getEnvironmentVariables(appID);
        if (getEnvironmentVariablesResult.isFailure()) {
            return;
        }

        try {
            const value = getEnvironmentVariablesResult.value!.value;
            const envVars = JSON.parse(value);
            setEnvVars(Object.getOwnPropertyNames(envVars).map(key => ({ key: key, value: envVars[key] })));
        } catch (err) {

        }


        setIsLoading(false);
    }

    const getEnvVarsAsString = (): string => {
        const value: Record<string, string> = {};

        envVars.forEach((variable) => {
            value[variable.key] = variable.value;
        });

        return JSON.stringify(value);
    }

    const createEnvironmentVariablesHandler = async () => {
        setIsLoading(true);

        const createEnvironmentVariablesResult = await createEnvironmentVariables(appID, getEnvVarsAsString());
        if (createEnvironmentVariablesResult.isFailure()) {
            return;
        }

        setIsLoading(false);
    }

    const updateEnvironmentVariablesHandler = async () => {
        setIsLoading(true);

        const updateEnvironmentVariablesResult = await updateEnvironmentVariables(appID, getEnvVarsAsString());
        if (updateEnvironmentVariablesResult.isFailure()) {
            return;
        }

        setIsLoading(false);
    }

    return {
        isLoading,
        envVars,
        setEnvVars,
        loadEnvironmentVariables,
        createEnvironmentVariablesHandler,
        updateEnvironmentVariablesHandler,
    };
}