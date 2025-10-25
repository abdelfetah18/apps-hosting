import type { Route } from "./+types/app_settings";

export default function AppSettings({ }: Route.ComponentProps) {
    return (
        <div className="w-full flex flex-col px-16 py-8 overflow-auto">
            <div className="w-full flex flex-col gap-2">
                <div className="flex flex-col gap-4">
                    <div className="flex items-start justify-between">
                        <div className="text-xl text-balance font-semibold hover:underline">Settings</div>
                    </div>
                </div>
            </div>
        </div>
    );
}
