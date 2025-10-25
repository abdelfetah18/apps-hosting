# Apps Hosting

**Apps Hosting** is an alternative to platforms like [Render](https://render.com/).
It provides one-click deployments — just provide a GitHub repository URL, and the app will be built, deployed, and assigned a custom domain automatically.

---

### Project Structure

* `docs/` – Documentation files
* `config/` – Configuration files, global environment variables, and TLS certificates *(not committed to Git)*
* `helm-chart/` – Helm charts used to deploy the system on Kubernetes
* `infrastructure/` – Tools and resources for deploying and managing the platform
* `internal-packages/` – Private packages used internally by the project
* `scripts/` – Utility scripts (e.g., Minikube setup)
* `src/` – Source code for all services

