# Building a Scalable and Responsive Code Execution System with Go

## Understanding the Problem

The core functionality of a code execution system is to allow users to submit code, execute it on a server, and return the results. This type of system is commonly found in online coding platforms, where users can practice and test their code without setting up a local development environment.

The penitervew's video on YouTube inspires this project: _Design a Code Execution System | System Design_. take a look at the video [here](https://www.youtube.com/watch?v=TOyD-5QgpuE)

### Some of the key requirements for such a system include:

- Functional Requirements:

  - Run any code submitted by a user
  - Return a result indicating whether the code executed successfully

- Non-Functional Requirements:

  - Low latency - users expect near-instant execution of their code
  - Isolation between user code executions to ensure security
  - Fault tolerance to handle failures gracefully

## Architecture Overview

We combined Docker and Kubernetes to achieve the desired level of isolation, scalability, and fault tolerance.

The high-level architecture will consist of the following components:

![Architecture Overview](/docs/Architecture.png)

- **API Server**: The API server will receive code execution requests from users and store them in a message queue for processing.

- **Pod manager**: The pod manager will be responsible for managing the lifecycle of pods in the Kubernetes cluster. It will run the code inside the containers, handle timeout, and monitor their status.

- **Standby pods**: The standby pods will be used to execute the code submitted by users. These pods will be pre-created and waiting for incoming requests.

- **Running pods**: The running pods are the pods that are currently executing user code, and will be destroyed after the code execution is complete.

## Running the Code Execution System

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
  go run code-execution-service/main.go
```
