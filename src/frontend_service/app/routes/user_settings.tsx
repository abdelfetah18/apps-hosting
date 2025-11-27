import { useFetcher } from "react-router";
import type { Route } from "./+types/user_settings";
import Spinner from "~/components/spinner";
import Iconify from "~/components/Iconify";
import { DEFAULT_PROFILE_PICTURE } from "~/consts";

export default function UserSettings({ }: Route.ComponentProps) {
    const fetcher = useFetcher();
    const isDraft = false;

    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                <div className="flex flex-col gap-8">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">User settings</div>
                    </div>
                    <fetcher.Form method="POST" className="w-full flex flex-col gap-4">
                        {/* Name */}
                        <div className="flex">
                            <div className="w-1/3 flex flex-col gap-1 text-black">
                                <div>Username</div>
                                <div className="text-sm text-gray-500">A unique username for your account.</div>
                            </div>
                            <div className="h-fit grow flex flex-col gap-2">
                                <input
                                    name="username"
                                    type="text"
                                    placeholder="Username"
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.name ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                            </div>
                        </div>

                        {/* Email */}
                        <div className="flex">
                            <div className="w-1/3 flex flex-col gap-1 text-black">
                                <div>Email</div>
                                <div className="text-sm text-gray-500">An Email for your account.</div>
                            </div>
                            <div className="h-fit grow flex flex-col gap-2">
                                <input
                                    name="email"
                                    type="text"
                                    placeholder="Email"
                                    className={`w-full border rounded-lg px-4 py-2 text-sm ${fetcher.data?.name ? "border-red-300" : "border-gray-300"}`}
                                />
                                <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                            </div>
                        </div>


                        {/* Plan */}
                        <div className="flex">
                            <div className="w-1/3 flex flex-col gap-1 text-black">
                                <div>Profile Image</div>
                                <div className="text-sm text-gray-500">A Profile picture for your account.</div>
                            </div>
                            <div className="h-fit grow flex flex-col gap-2">
                                <div className="w-20 h-20 rounded-full bg-gray-200 relative">
                                    <img src={DEFAULT_PROFILE_PICTURE} alt="user profile picture" className="w-full h-full object-cover rounded-full select-none" />
                                    <div className="absolute bottom-0 right-0 bg-blue-500 hover:bg-blue-600 cursor-pointer rounded-full text-white p-2">
                                        <Iconify icon="bxs:camera" size={16} />
                                    </div>
                                </div>
                                <div className="text-red-600 text-xs">{fetcher.data?.name}</div>
                            </div>
                        </div>

                        <div className="w-full flex items-center justify-end">
                            {
                                isDraft && (
                                    <div className="flex items-center gap-2">
                                        <div className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none border border-purple-700 text-purple-700 text-center rounded-full py-1 font-semibold">Cancel</div>
                                        {
                                            fetcher.state == "idle" ? (

                                                <button
                                                    type="submit"
                                                    className="w-30 cursor-pointer hover:bg-purple-500 active:scale-105 duration-300 select-none bg-purple-700 text-white text-center rounded-full py-1 font-semibold"
                                                >
                                                    Save
                                                </button>
                                            ) : (
                                                <button
                                                    type="button"
                                                    className="w-30 flex items-center justify-center cursor-pointer select-none py-2"
                                                >
                                                    <Spinner />
                                                </button>
                                            )
                                        }
                                    </div>
                                )
                            }
                        </div>
                    </fetcher.Form>

                </div>
            </div>
            <div>
                <div></div>
            </div>
        </div>
    );
}
