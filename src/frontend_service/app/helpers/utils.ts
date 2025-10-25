export function toActivityList(builds: Build[], deployments: Deployment[]): Activity[] {
    const buildsActivity: Activity[] = builds.map(build => ({ ...build, type: "build" }));
    const deploymentsActivity: Activity[] = deployments.map(deployment => ({ ...deployment, type: "deployment" }));

    const allActivities = [...buildsActivity, ...deploymentsActivity];
    
    return allActivities.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
}
