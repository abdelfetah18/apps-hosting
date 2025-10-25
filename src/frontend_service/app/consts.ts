export const DEFAULT_PROFILE_PICTURE = "/default_profile_picture.png";
export const RUNTIMES: Runtime[] = ["NodeJS"];
export const RUNTIMES_LOGOS: { [key in Runtime]: string; } = {
    NodeJS: "logos:nodejs-icon",
};

// export const APP_STATUS_ICONS: { [key in AppStatus]: string; } = {
//     build_failed: "material-symbols:error-outline",
//     building: "mdi:progress-clock",
//     deploy_failed: "material-symbols:error-outline",
//     deployed: "mdi:check-bold",
//     deploying: "mdi:progress-clock",
// };

// export const APP_STATUS_COLORS: { [key in AppStatus]: string; } = {
//     build_failed: "text-red-700",
//     building: "text-yellow-700",
//     deploy_failed: "text-red-700",
//     deployed: "text-green-700",
//     deploying: "text-yellow-700",
// };

export const DEPLOYMENT_STATUS_COLORS: { [key in DeploymentStatus]: string; } = {
    failed: "text-red-700",
    successed: "text-green-700",
    pending: "text-yellow-700",
};

export const DEPLOYMENT_STATUS_ICONS: { [key in DeploymentStatus]: string; } = {
    failed: "material-symbols:error-outline",
    successed: "mdi:check-bold",
    pending: "mdi:progress-clock",
};

export const BUILD_STATUS_COLORS: { [key in BuildStatus]: string; } = {
    failed: "text-red-700",
    successed: "text-green-700",
    pending: "text-yellow-700",
};

export const BUILD_STATUS_ICONS: { [key in BuildStatus]: string; } = {
    failed: "material-symbols:error-outline",
    successed: "mdi:check-bold",
    pending: "mdi:progress-clock",
};