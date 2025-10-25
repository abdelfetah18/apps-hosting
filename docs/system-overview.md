# Apps Hosting Overview

### System Architecture

* **Architecture Style:** Microservices-based
* **Database Strategy:** **Database-per-Service** pattern (each service owns its own data)
* **Authentication Approach:** **Edge-level authorization** (handled at the Gateway)
* **Service Communication:**

  * **NATS** for event-driven communication
  * **gRPC** for service-to-service communication
* **Tech Stack:**

    * All backend services are written in **GoLang**
    * Frontend is built using the **React Router** framework
    * **Kaniko** is used for building Docker images (chosen instead of Docker-in-Docker for security reasons)
    * **Postgres** as the database for all services (Database-per-Service pattern)

---

### Core Services

* User Service (`user_service`)
* Project Service (`project_service`)
* Application Service (`app_service`)
* Build Service (`build_service`)
* Deployment Service (`deploy_service`)
* Logging Service (`log_service`)
* Gateway Service (`gateway_service`)
* Frontend Service (`frontend_service`)

---

### User Service

Responsible for:

1. User Authentication
2. User Profile Management
3. GitHub Authorization

---

### Project Service

Responsible for:

1. Project Management

---

### Application Service

Responsible for:

1. Application Management

---

### Build Service

Responsible for building the user's application from source code.

Build Pipeline:

* Clone the user’s repository
* Add the appropriate Dockerfile based on the selected runtime
* Build and publish the Docker image to the registry using **Kaniko**
* Publish a message once the build is completed

---

### Deployment Service

Responsible for deploying the application to Kubernetes.

Deployment Pipeline:

* Create Kubernetes **Deployment**, **Service**, and **ConfigMap** resources
* Create an **Ingress** resource to expose the app using a custom domain
* Publish a message once the deployment is completed

---

### Logging Service

Responsible for:

1. Reading and providing user application logs

---

### Gateway Service

Acts as the single entry point for all external requests.

* Implements the **Gateway Pattern** for the microservices architecture
* All requests pass through this service
* Handles authentication at the edge level

---

### Frontend Service

* A React Router application that serves as the UI for the system
