import axios, { type AxiosInstance } from "axios";
import { getAccessToken } from "./user_service";

export function getAxiosInstance(): AxiosInstance {
    if (axios.defaults.headers.common.Authorization) {
        return axios;
    }

    const token = getAccessToken();
    axios.defaults.headers.common.Authorization = token;
    return axios;
}