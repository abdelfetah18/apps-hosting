import { AxiosError } from "axios";
import Result from "~/helpers/Result";
import { getAxiosInstance } from "./api_service";

const axios = getAxiosInstance();
axios.defaults.baseURL = "https://api.apps-hosting.com";

export async function listProjects(): Promise<Result<string, Project[]>> {
    try {
        const response = await axios.get("/projects");
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

export async function getProjectById(projectId: string): Promise<Result<string, Project>> {
    try {
        const response = await axios.get(`/projects/${projectId}`);
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

export async function createProject(createProjectForm: CreateProjectForm): Promise<Result<string, Project>> {
    try {
        const response = await axios.post<HttpResponse<Project>>("/projects/create", createProjectForm);
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

export async function deleteProjectById(projectId: string): Promise<Result<string, Project>> {
    try {
        const response = await axios.delete(`/projects/${projectId}/delete`);
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
