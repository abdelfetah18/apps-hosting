import { useSearchParams } from "react-router";
import type { Route } from "./+types/github_callback";
import { useEffect } from "react";
import { userGithubCallback } from "~/services/user_service";

export default function GithubCallback({ params }: Route.ComponentProps) {
    const [searchParams] = useSearchParams();

    const code = searchParams.get("code") || "";
    const setupAction = searchParams.get("setup_action");

    const next = async () => {
        const result = await userGithubCallback(code);
        if (result.isFailure()) {
            alert(result.error);
        } else {
            alert(result.value);
        }
    }

    useEffect(() => {
        if (setupAction != "update") {
            next();
        }

        window.opener.postMessage("ping");
        window.close();
    }, []);

    return (
        <div className="w-full flex flex-col gap-4 p-8 ">
            <div onClick={next} className="bg-purple-700 w-96 py-2 rounded-full text-white font-bold text-center cursor-pointer active:scale-105 select-none hover:bg-purple-600 duration-300">Auth</div>
        </div>
    );
}
