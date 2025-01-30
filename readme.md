# Building a Scalable and Responsive Code Execution System with Go

## Understanding the Problem

The core functionality of a code execution system is to allow users to submit code, execute it on a server, and return the results. This type of system is commonly found in online coding platforms, where users can practice and test their code without setting up a local development environment.

The penitervew's video on YouTube inspires this project: _Design a Code Execution System | System Design_. take a look at the video [here](https://www.youtube.com/watch?v=TOyD-5QgpuE)

## Overview

The **Code Execution System** provides an efficient and scalable solution for executing code snippets in multiple programming languages within a Kubernetes-based infrastructure. It leverages a **task queue**, a **scheduler**, and dynamically managed **Kubernetes pods** to handle concurrent execution requests while ensuring security, scalability, and observability.

## 1. Requirements

### Functional Requirements

- **Task Submission**: Users submit code snippets for execution via an API.
- **Multi-Language Support**: Supports multiple programming languages (e.g., Python, Java, C).
- **Result Retrieval**: Users can retrieve execution results using a unique task ID.
- **Concurrency**: Executes multiple tasks concurrently for optimal resource utilization.
- **Extensibility**: Easily extendable to support additional languages and execution environments.

### Non-Functional Requirements

- **Scalability**: Efficiently handles a growing number of concurrent tasks.
- **High Availability**: Ensures reliable execution with minimal downtime.
- **Security**: Isolates execution environments to prevent unauthorized access.
- **Performance**: Low-latency scheduling and execution.
- **Maintainability**: Modular architecture for easy updates and debugging.
- **Observability**: Integrates logging and monitoring tools for tracking task execution and system health.

---

## 2. System Workflow

1. **Task Submission**

   - Users submit a code execution request via the `/submit` endpoint.
   - The request includes the programming language and code.

2. **Task Queuing**

   - The API layer publishes the task to a **RabbitMQ queue**.
   - Each programming language has its own queue for efficient handling.

3. **Task Scheduling**

   - The **Scheduler** consumes tasks from the queue.
   - It requests a **standby pod** from the **Pod Manager**.

4. **Code Execution**

   - The scheduler assigns the task to an available Kubernetes pod.
   - The pod executes the code in an isolated environment.
   - Execution results (stdout, stderr) are captured.

5. **Result Storage**
   - Results are stored in a **PostgreSQL database**.
   - Users retrieve results via the `/result/:task_id` endpoint.

---

## 3. Key Features

- **Multi-Language Execution**: Easily configurable to add new programming languages.
- **RabbitMQ Task Queue**: Asynchronous task management ensures efficient execution.
- **Kubernetes Integration**: Automatic pod scaling based on workload demand.
- **Pod Pooling Mechanism**: Standby pods are pre-warmed for lower execution latency.
- **RESTful API**: User-friendly interface for submitting and retrieving tasks.
- **Task Observability**: Tracks task lifecycle for debugging and analytics.

---

## 4. Architecture Components

We combined Docker and Kubernetes to achieve the desired level of isolation, scalability, and fault tolerance.

The high-level architecture will consist of the following components:

![Architecture Overview](/docs/architecture.png)

### API Layer

- Accepts execution requests and task result queries.
- Publishes tasks to the queue.
- Built using the **Gin** framework (Go).

### Task Queue (RabbitMQ)

- Stores pending execution tasks.
- Uses **language-specific queues** for better performance.

### Scheduler

- Consumes tasks from the queue.
- Assigns tasks to available **standby pods**.
- Implements **callbacks** to monitor task execution.

### Pod Manager

- Maintains a **pool of standby pods**.
- Allocates pods dynamically to optimize resource utilization.
- Interacts with the **Kubernetes API**.

### Execution Pods

- Runs user-submitted code in an isolated environment.
- Returns execution results to the **Scheduler**.

### Storage Layer (PostgreSQL)

- Stores task metadata and execution results.
- Ensures persistence for task retrieval.

---

## 7. Future Enhancements

- **Persistent Database**: Extend PostgreSQL storage with **Redis** caching for faster retrieval.
- **Advanced Scheduling**: Implement **priority-based scheduling**.
- **Monitoring & Logging**: Integrate **Prometheus & Grafana** for system observability.
- **Enhanced Security**: Introduce **sandboxing & resource limits** for better isolation.
- **Extended Language Support**: Add more compilers and interpreters.

## 8. Running the Code Execution System

To run the code execution system, follow these steps:

1. Install Docker and Kubernetes on your machine.

2. Create a Kubernetes cluster using Minikube or any other Kubernetes provider.

```bash
  minikube start
```

3. create a namespace for the code execution system

```bash
  kubectl create namespace code-exec-system
```

4. Deploy the standby pods to the Kubernetes cluster.

```bash
  kubectl apply -f k8s-config/standby-pod-deployment.yaml
```

5. run the API server using air or go run

```bash
  cd code-execution-service
  go run main.go
```
