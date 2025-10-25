import Result from "~/helpers/Result";
import { getAxiosInstance } from "./api_service";
import type { AxiosError } from "axios";

export function getAccessToken(): string | null {
    return localStorage.getItem("token");
}

const axios = getAxiosInstance();
axios.defaults.baseURL = "https://api.apps-hosting.com";

export async function getUserGithubRepositories(): Promise<Result<string, GithubRepository[]>> {
    try {
        const response = await axios.get("/user/github/repositories");
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

export async function userGithubCallback(code: string): Promise<Result<string, string>> {
    try {
        const response = await axios.get(`/user/github/callback?code=${code}`);
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.message!);
    } catch (error) {
        const err = error as AxiosError<any>;
        console.error(err.response?.data.message);
        return Result.failure(err.response?.data.message);
    }
}

