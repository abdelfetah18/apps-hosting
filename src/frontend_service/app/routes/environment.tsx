import Iconify from "~/components/Iconify";
import type { Route } from "./+types/environment";
import { useEffect } from "react";
import useEnvironmentVariables from "~/hooks/useEnvironmentVariables";

export default function Environment({ params }: Route.ComponentProps) {
    const { envVars, setEnvVars, isLoading, loadEnvironmentVariables, updateEnvironmentVariablesHandler } = useEnvironmentVariables(params.app_id);

    useEffect(() => {
        loadEnvironmentVariables();
    }, []);

    const handleAddEnvVar = () => {
        setEnvVars([...envVars, { key: "", value: "" }]);
    };

    const handleEnvVarChange = (index: number, field: "key" | "value", value: string) => {
        const updated = [...envVars];
        updated[index][field] = value;
        setEnvVars(updated);
    };

    const handleRemoveEnvVar = (index: number) => {
        setEnvVars(envVars.filter((_, i) => i !== index));
    };

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                <div className="flex flex-col gap-8">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">Environment</div>
                    </div>

                    <div className="w-full flex flex-col gap-4">
                        <div className="flex flex-col gap-1">
                            <div className="text-black font-semibold">Environment Variables</div>
                            <div className="text-gray-600 text-sm whitespace-pre-line">{"Set environment-specific config and secrets (such as API keys), then read those\nvalues from your code."}</div>
                        </div>
                        <div className="w-full flex flex-col gap-2">
                            {envVars.map((env, index) => (
                                <div key={index} className="flex gap-2 items-center">
                                    <input
                                        name="env_key"
                                        placeholder="Key"
                                        value={env.key}
                                        onChange={(e) => handleEnvVarChange(index, "key", e.target.value)}
                                        className="flex-1 border border-gray-300 rounded-lg px-4 py-2 text-sm"
                                    />
                                    <input
                                        name="env_value"
                                        placeholder="Value"
                                        value={env.value}
                                        onChange={(e) => handleEnvVarChange(index, "value", e.target.value)}
                                        className="flex-1 border border-gray-300 rounded-lg px-4 py-2 text-sm"
                                    />
                                    {envVars.length > 1 && (
                                        <button
                                            type="button"
                                            onClick={() => handleRemoveEnvVar(index)}
                                            className="text-red-500 cursor-pointer"
                                        >
                                            <Iconify icon="material-symbols:delete-outline" size={16} />
                                        </button>
                                    )}
                                </div>
                            ))}
                            <button
                                type="button"
                                onClick={handleAddEnvVar}
                                className="flex items-center gap-1 text-sm text-blue-600 hover:underline w-fit mt-1 cursor-pointer"
                            >
                                <Iconify icon="material-symbols:add-rounded" /> Add Variable
                            </button>
                        </div>
                        <div className="w-full flex flex-col items-end">
                            <button
                                type="button"
                                onClick={updateEnvironmentVariablesHandler}
                                className="w-96 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none bg-purple-700 text-white text-center rounded-full py-2 font-semibold"
                            >
                                Save
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
