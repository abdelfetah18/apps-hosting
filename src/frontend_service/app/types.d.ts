interface HttpResponse<T> {
    status: string;
    data?: T;
    message?: string;
}

interface UserCredentials {
    email: string;
    password: string;
}

interface CreateUserForm {
    username: string;
    email: string;
    password: string;
}

interface User {
    id: string;
    username: string;
    email: string;
    github_app_installed: string;
    github_access_token: string;
    github_refresh_token: string;
    created_at: string;
}

interface UserSession {
    access_token: string;
    user: User;
}

type Runtime = "NodeJS";
interface App {
    id: string;
    runtime: Runtime;
    name: string;
    repo_url: string;
    user_id: string;
    start_cmd: string;
    build_cmd: string;
    domain_name: string;
    created_at: string;
    build: Build;
}

interface EnvironmentVariables {
    id: string;
    app_id: string;
    value: string;
    created_at: string;
}

type GitProvider = "github"
interface GitRepository {
    id: string;
    clone_url: string;
    provider: GitProvider;
    is_private: boolean;
    created_at: string;
}

type BuildStatus = "successed" | "failed" | "pending";
interface Build {
    id: string;
    app_id: string;
    image_url: string;
    status: BuildStatus;
    commit_hash: string;
    created_at: string;
}

type DeploymentStatus = "successed" | "failed" | "pending";
interface Deployment {
    id: string;
    build_id: string;
    app_id: string;
    status: DeploymentStatus;
    created_at: string;
}

interface CreateAppForm {
    name: string;
    runtime: string;
    repo_url: string;
    start_cmd: string;
    build_cmd: string;
    git_repository: {
        clone_url: string;
        is_private: boolean;
        provider: GitProvider;
    };
}

interface UpdateAppForm {
    name: string;
    start_cmd: string;
    build_cmd: string;
}

interface Environment {
    variables: Record<string, string>;
}

type ToastMessageType = "error" | "success" | "info";

interface ToastMessage {
    id: number;
    message: string;
    type: MessageType;
};

interface useToastReturn {
    messages: ToastMessage[];
    alertError: (message: string, duration?: number) => void;
    alertSuccess: (message: string, duration?: number) => void;
    alertInfo: (message: string, duration?: number) => void;
}

type BuildActivity = Build & {
    type: "build";
};

type DeploymentActivity = Deployment & {
    type: "deployment";
};

type Activity = BuildActivity | DeploymentActivity;


interface Project {
    id: string;
    name: string;
    user_id: string;
    created_at: string;
}

interface CreateProjectForm {
    name: string;
}

interface UpdateProjectForm {
    name: string;
}

interface GithubRepository {
    id: string;
    name: string;
    description: string;
    default_branch: string;
    git_url: string;
    clone_url: string;
    url: string;
    visibility: string;
}

type GetUserProjectsResponse = Array<{ id: string; name: string; apps_count: number; created_at: string; }>; 