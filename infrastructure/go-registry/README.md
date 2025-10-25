# Go Registry

**Go Registry** is a lightweight **Golang module proxy server** — similar to **Google Artifact Registry**, but self-hosted.
It lets you host, store, and serve Go modules locally or within your organization using **MinIO** for storage.

---

## 1. Environment Setup

Create a `.env` file in the project root with your MinIO configuration:

```bash
MINIO_URL=127.0.0.1:9000
ACCESS_KEY_ID=minioadmin
SECRET_ACCESS_KEY=minioadmin
BUCKET_NAME=golang-registry
```

Make sure your MinIO server is running and accessible at the URL you set above.

---

## 2. Configure Go to Use Your Registry

Tell Go to use your local registry first, then fall back to the public proxy and source:

```bash
export GOPROXY=http://localhost:8080,https://proxy.golang.org,direct
```

If you’re using `localhost`, allow insecure connections:

```bash
export GOINSECURE=localhost:8080
```

---

## 3. Exclude Private Modules from Public Checksum Verification

If you host private modules, exclude them from the public checksum database:

```bash
export GONOSUMDB=MODULE_PATH_REGEX
```

Examples:

Exclude a single module:

```bash
export GONOSUMDB=example.com/foo
```

Exclude all modules under a domain:

```bash
export GONOSUMDB=example.com/*
```

---

✅ **You’re all set!**
Run your Go Registry server and start publishing or downloading modules from your local MinIO-backed proxy.
