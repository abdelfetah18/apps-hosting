import { Link, redirect, useFetcher } from "react-router";
import type { Route } from "./+types/sign_in";
import { signIn } from "~/services/auth_service";
import ToastContext from "~/contexts/ToastContext";
import { useContext, useEffect } from "react";

export async function clientAction({ request }: Route.ClientActionArgs) {
    const formData = await request.formData();

    const email = formData.get("email")?.toString() || "";
    const password = formData.get("password")?.toString() || "";

    const errors: Record<string, string> = {};

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

    const result = await signIn({ email, password });
    if (result.isSuccess()) {
        localStorage.setItem("token", result.value?.access_token || "");
        return redirect("/");
    }

    errors.api = result.error!;
    return errors;
}

export default function SignIn({ }: Route.ComponentProps) {
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
                    <div className="text-3xl font-semibold">Welcome Back to Apps Hosting</div>
                    <div className="text-base text-gray-500">Please Enter your details!</div>
                </div>
                <fetcher.Form method="POST" className="w-full flex flex-col items-center gap-6">
                    <div className="w-full flex flex-col items-center gap-2">
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
                            <button className="w-full text-white bg-purple-700 hover:bg-purple-600 rounded-full py-2 font-semibold cursor-pointer" type="submit">Sign In</button>
                        )
                    }
                    <div className="w-full flex items-center gap-1 text-sm">
                        <div>Need an account?</div>
                        <Link to="/sign_up">
                            <div className="font-semibold text-purple-700">Sign Up</div>
                        </Link>
                    </div>
                </fetcher.Form>
            </div>
        </div>
    )
}