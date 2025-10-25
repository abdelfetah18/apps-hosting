import { AxiosError } from "axios";
import Result from "~/helpers/Result";
import { getAxiosInstance } from "./api_service";

const axios = getAxiosInstance();
axios.defaults.baseURL = "https://api.apps-hosting.com";

export async function listApps(projectId: string): Promise<Result<string, App[]>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function getAppById(projectId: string, appId: string): Promise<Result<string, App>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appId}`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function createApp(projectId: string, createAppForm: CreateAppForm): Promise<Result<string, App>> {
    try {
        const response = await axios.post<HttpResponse<App>>(`/projects/${projectId}/apps/create`, createAppForm);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function reDeployAppById(projectId: string, appId: string): Promise<Result<string, string>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appId}/redeploy`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.message);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function getEnvironmentVariables(projectId: string, appId: string): Promise<Result<string, EnvironmentVariables>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appId}/environment_variables`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function createEnvironmentVariables(projectId: string, appId: string, value: string): Promise<Result<string, EnvironmentVariables>> {
    try {
        const response = await axios.post(`/projects/${projectId}/apps/${appId}/environment_variables/create`, { value });
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function updateEnvironmentVariables(projectId: string, appId: string, value: string): Promise<Result<string, EnvironmentVariables>> {
    try {
        const response = await axios.patch(`/projects/${projectId}/apps/${appId}/environment_variables/update`, { value });
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function getAppLogs(projectId: string, appId: string): Promise<Result<string, string>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appId}/logs`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }

}

export async function deleteAppById(projectId: string, appID: string): Promise<Result<string, App>> {
    try {
        const response = await axios.delete(`/projects/${projectId}/apps/${appID}/delete`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }
}

export async function listDeployments(projectId: string, appID: string): Promise<Result<string, Deployment[]>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appID}/deployments`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message || err.response?.data);
    }
}

export async function listBuilds(projectId: string, appID: string): Promise<Result<string, Build[]>> {
    try {
        const response = await axios.get(`/projects/${projectId}/apps/${appID}/builds`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message || err.response?.data);
    }

}