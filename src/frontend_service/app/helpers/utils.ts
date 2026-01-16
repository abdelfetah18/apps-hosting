export function toActivityList(builds: Build[], deployments: Deployment[]): Activity[] {
    const buildsActivity: Activity[] = builds.map(build => ({ ...build, type: "build" }));
    const deploymentsActivity: Activity[] = deployments.map(deployment => ({ ...deployment, type: "deployment" }));

    const allActivities = [...buildsActivity, ...deploymentsActivity];

    return allActivities.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
}

/**
 * Checks whether the query exists within the given text.
 * Returns true if the query is empty.
 */
export function textContainsQuery(text: string, query: string): boolean {
    const normalize = (input: string): string => {
        return input
            .toLowerCase()
            .replace(/[-_]+/g, ' ')
            .replace(/\s+/g, ' ')
            .trim();
    }

    return normalize(text).includes(normalize(query));
}
