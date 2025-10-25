import { Link, redirect, useFetcher } from "react-router";
import type { Route } from "./+types/sign_up";
import { signUp } from "~/services/auth_service";
import { useContext, useEffect } from "react";
import ToastContext from "~/contexts/ToastContext";

export async function clientAction({ request }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const username = formData.get("username")?.toString() || "";
    const email = formData.get("email")?.toString() || "";
    const password = formData.get("password")?.toString() || "";

    const errors: Record<string, string> = {};

    if (username.length == 0) {
        errors.username = "Username is required.";
    }

    if (email.length == 0) {
        errors.email = "Email is required.";
    } else if (!email.match(/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/g)) {
        errors.email = "Please enter a valid email address.";
    }

    if (password.length == 0) {
        errors.password = "Password is required.";
    }

    if (Object.keys(errors).length > 0) {
        return errors;
    }

    const result = await signUp({ username, email, password });
    if (result.isSuccess()) {
        return redirect("/sign_in");
    }

    errors.api = result.error!;
    return errors;
}

export default function SignUp({ actionData }: Route.ComponentProps) {
    const toastManager = useContext(ToastContext);
    const fetcher = useFetcher();

    useEffect(() => {
        if (fetcher.data?.api != undefined) {
            toastManager.alertError(fetcher.data.api);
        }
    }, [fetcher.data]);

    return (
        <div className="w-full h-screen flex flex-col items-center">
            <div className="w-2/5 h-full flex flex-col gap-8 items-center justify-center p-12">
                <div className="w-full flex flex-col gap-1">
                    <div className="text-3xl font-semibold">Welcome to Apps Hosting</div>
                    <div className="text-base text-gray-500">Please Enter your details!</div>
                </div>
                <fetcher.Form method="POST" className="w-full flex flex-col items-center gap-8">
                    <div className="w-full flex flex-col items-center gap-2">
                        <div className="w-full flex flex-col gap-2">
                            <div className="text-sm">Username:</div>
                            <div className="h-fit flex-grow flex flex-col gap-2">
                                <input
                                    name="username"
                                    type="text"
                                    placeholder="Type your username"
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.username ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.username}</div>
                            </div>
                        </div>
                        <div className="w-full flex flex-col gap-2">
                            <div className="text-sm">Email:</div>
                            <div className="h-fit flex-grow flex flex-col gap-2">
                                <input
                                    name="email"
                                    type="text"
                                    placeholder="Type your email"
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.email ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.email}</div>
                            </div>
                        </div>
                        <div className="w-full flex flex-col gap-2">
                            <div className="text-sm">Password:</div>
                            <div className="h-fit flex-grow flex flex-col gap-2">
                                <input
                                    name="password"
                                    type="password"
                                    placeholder="Type your password"
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.password ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.password}</div>
                            </div>
                        </div>
                    </div>
                    {
                        fetcher.state == "submitting" ? (
                            <div>Loading...</div>
                        ) : (
                            <button className="w-full text-white bg-purple-700 hover:bg-purple-600 rounded-full py-2 font-semibold cursor-pointer" type="submit">Sign Up</button>
                        )
                    }
                </fetcher.Form>

                <div className="w-full flex items-center gap-1 text-sm">
                    <div>Already have an account?</div>
                    <Link to="/sign_in">
                        <div className="font-semibold text-purple-700">Sign In</div>
                    </Link>
                </div>
            </div>
        </div>
    )
}