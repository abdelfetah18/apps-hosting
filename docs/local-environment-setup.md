## Setup Guide: Apps Hosting in a Local Environment

### 1. Install Prerequisites

Ensure the following are installed before proceeding:

* [Minikube](https://minikube.sigs.k8s.io/docs/start)
* [Docker](https://www.docker.com/)
* [Helm](https://helm.sh/)
* [Mkcert](https://github.com/FiloSottile/mkcert)
* [GoLang](https://go.dev/)
* [GRPC](https://grpc.io/) and [Protobuf](https://protobuf.dev/)
* [NodeJS](http://nodejs.org/)
* A local DNS Server (to simulate real-world domain names)
* [Postgres](https://www.postgresql.org/)

---

### 2. Database Setup

This project uses a **microservices** architecture — meaning *each service has its own database*.
For now, database creation is manual (automation will be added later).

Create **five PostgreSQL databases** with the following names:

* `build_service`
* `app_service`
* `user_service`
* `deploy_service`
* `project_service`

> No tables need to be created — microservices will handle that automatically.

---

### 3. Add Database Connection Strings

For each microservice, create a `.env` file inside:

```
src/<service-name>/.env
```

Add:

```bash
DATABASE_CONNECTION_STRING=postgresql://[user[:password]@][host][:port][/dbname][?param1=value1&param2=value2...]
```

**Important:**
The DB host must be accessible *from inside Minikube*.
`localhost` will **not** work — use your Wi-Fi IP or an external cloud DB (e.g., [Neon](https://neon.com/)).

---

### 4. TLS Certificate Setup (Mkcert)

Generate TLS certificates:

```bash
mkcert apps-hosting.com
mkcert *.apps-hosting.com
```

Store the certificates in: `config/tls/`
This directory should contain:

* `config/tls/apps-hosting.com.pem`
* `config/tls/apps-hosting.com-key.pem`
* `config/tls/_wildcard.apps-hosting.com.pem`
* `config/tls/_wildcard.apps-hosting.com-key.pem`

---

### 5. Go Artifact Registry Setup

Inside `infrastructure/go-registry`, create the following files:

**`.env`**

```
MINIO_ENDPOINT=minio.apps-hosting.com
MINIO_ID=minioadmin
MINIO_SECRET=minioadmin
MINIO_TOKEN=
MINIO_BUCKET_NAME=golang-registry
```

**`.server.env`**

```
MINIO_ENDPOINT=object-storage-minio:9000
MINIO_ID=minioadmin
MINIO_SECRET=minioadmin
MINIO_TOKEN=
MINIO_BUCKET_NAME=golang-registry
```

---

### 6. Minikube Setup

Run the setup script:

```bash
scripts/minikube/setup.sh
```

This configures Minikube with Docker & Go artifact registries and sets up ingress with TLS secrets.

---

### 7. DNS Configuration

Point your local DNS to the Minikube IP:

```bash
minikube ip
```

Configure:

```
apps-hosting.com       -> <minikube-ip>
*.apps-hosting.com     -> <minikube-ip>
```

---

### 8. Upload Internal Go Packages

Publish internal Golang packages to the registry:

```bash
scripts/publish-internl-packages.sh
```

---

### 9. Global Environment Configuration

Create `config/.global.env` with:

```
PORT=8080
NATS_URL=nats
REGISTRY_URL=192.168.49.2:5000/
USER_SERVICE=user-service:8080
APP_SERVICE=app-service:8080
BUILD_SERVICE=build-service:8080
DEPLOY_SERVICE=deploy-service:8080
LOG_SERVICE=log-service:8080
PROJECT_SERVICE=project-service:8080
```

> `REGISTRY_URL` must match the Minikube IP:

```bash
minikube ip
```

---

### 10. Setup Environment Variables for Services

Add to the `user_service` `.env` file:

```
JWT_SECRET=my_secret_key
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
```

You can obtain `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` by creating your own GitHub App:
[https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)

For the other services (`gateway_service` and `log_service`), create an empty `.env` file.

---

### 11. Deploy All Services

Run the script to deploy everything (including NATS):

```bash
scripts/minikube/deploy.sh
```

