# Distributed Code Runner

## Overview

Distributed Code Runner is a system designed to distribute and execute Python code across multiple worker nodes. The system consists of a master node that manages job submissions and worker nodes that execute the jobs. The master node provides an API for submitting jobs and checking the status of workers.

## Features

- **Job Submission**: Submit Python code to be executed by worker nodes.
- **Job Distribution**: Distribute jobs to available worker nodes.
- **Job Execution**: Execute Python code on worker nodes.
- **Status Reporting**: Report the status of jobs and worker nodes.
- **Scalability**: Easily add more worker nodes to handle increased load.

## Use Cases

- **Distributed Computing**: Distribute computational tasks across multiple nodes to improve performance and efficiency.
- **Batch Processing**: Submit and process multiple jobs in parallel.
- **Load Balancing**: Distribute jobs evenly across worker nodes to balance the load.
- **Monitoring**: Monitor the status of jobs and worker nodes in real-time.

## Architecture

The system architecture consists of the following components:

- **Master Node**: Manages job submissions and distributes them to worker nodes.
- **Worker Nodes**: Execute the submitted Python code and report their status back to the master node.

### Architecture Diagram

![Architecture Diagram](./diagrams/architecture.png)

## Flow

1. **Job Submission**: Users submit Python code to the master node via an API.
2. **Job Distribution**: The master node distributes the job to an available worker node.
3. **Job Execution**: The worker node executes the Python code.
4. **Status Reporting**: The worker node reports the status of the job back to the master node.



### Prerequisites

- Node.js
- MongoDB
- Golang
