import { AxiosError, type AxiosResponse } from "axios";
import Result from "~/helpers/Result";
import { getAxiosInstance } from "./api_service";

const axios = getAxiosInstance();
axios.defaults.baseURL = "https://api.apps-hosting.com";

export async function signIn(userCredentials: UserCredentials): Promise<Result<string, UserSession>> {
    try {
        const response = await axios.post<HttpResponse<UserSession>>("/user/sign_in", userCredentials);
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

export async function signUp(createUserForm: CreateUserForm): Promise<Result<string, User>> {
    try {
        const response = await axios.post<HttpResponse<User>>("/user/sign_up", createUserForm);
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

export async function auth(accessToken: string): Promise<Result<string, User>> {
    try {
        const response = await axios.get<HttpResponse<User>>("/user/auth", { headers: { Authorization: accessToken } });
        if (response.data.status == "error") {
            return Result.failure(response.data.message!);
        }
        return Result.success(response.data.data!);
    } catch (error) {
        console.error(error);
        return Result.failure("Something went wrong");
    }

}