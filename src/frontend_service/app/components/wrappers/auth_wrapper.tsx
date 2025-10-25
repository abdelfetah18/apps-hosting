import { Outlet, redirect } from "react-router";
import { auth } from "~/services/auth_service";
import type { Route } from "./+types/auth_wrapper";
import axios from "axios";

export async function clientLoader() {
    const accessToken = window.localStorage.getItem("token");
    axios.defaults.headers.authorization = accessToken;

    if (!accessToken) {
        throw redirect("/sign_in");
    }

    const result = await auth(accessToken);
    if (result.isFailure()) {
        throw redirect("/sign_in");
    }

    return { access_token: accessToken, user: result.value! } as UserSession;
}

export default function AuthWrapper({ loaderData }: Route.ComponentProps) {
    return (
        <div className="w-full h-screen overflow-auto">
            <Outlet context={loaderData} />
        </div>
    );
}