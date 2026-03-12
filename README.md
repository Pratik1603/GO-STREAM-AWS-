# 🎬 GoStream — Movie Streaming App

<div align="center">

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![JavaScript](https://img.shields.io/badge/JavaScript-F7DF1E?style=for-the-badge&logo=javascript&logoColor=black)
![AWS Lambda](https://img.shields.io/badge/AWS_Lambda-FF9900?style=for-the-badge&logo=aws-lambda&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Vercel](https://img.shields.io/badge/Vercel-000000?style=for-the-badge&logo=vercel&logoColor=white)

A high-performance movie streaming application powered by a **Go backend** running on **AWS Lambda** — serverless, scalable, and fast.

🌐 **Live Demo:** [go-stream-aws.vercel.app](https://go-stream-aws.vercel.app)

</div>

---

## ✨ Features

- 🎥 **Smooth Video Streaming** — Efficient media delivery using Go's concurrency model
- ⚡ **Serverless Backend** — Go functions deployed on AWS Lambda for auto-scaling and zero server management
- 🐳 **Docker Support** — Run the full stack locally with a single command
- ☁️ **AWS Powered** — Leverages Lambda + S3 for scalable, cost-efficient infrastructure
- 🖥️ **Modern Frontend** — Clean, responsive UI for a seamless viewing experience

---

## 🏗️ Architecture

```
┌─────────────────┐         ┌──────────────────────────┐
│                 │  HTTP   │                          │
│  React/JS       │ ──────► │   AWS API Gateway        │
│  Client         │         │         │                │
│  (Vercel)       │ ◄────── │         ▼                │
│                 │  Stream │   AWS Lambda (Go)        │
└─────────────────┘         │         │                │
                             │         ▼                │
                             │      AWS S3              │
                             │   (Movie Storage)        │
                             └──────────────────────────┘
```

### Local Development (Docker)

```
┌─────────────────┐         ┌──────────────────────┐
│  React/JS       │ ──────► │  Go Server           │
│  Client         │ ◄────── │  (Docker Container)  │
└─────────────────┘         └──────────────────────┘
```

---

## 🗂️ Project Structure

```
GO-STREAM-AWS/
├── client/              # JavaScript frontend (deployed on Vercel)
│   └── ...
├── server/              # Go backend
│   ├── Dockerfile       # Docker config for local development
│   ├── main.go          # Lambda handler entry point
│   └── ...
└── .gitignore
```

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) `>= 1.21`
- [Node.js](https://nodejs.org/) `>= 18`
- [Docker](https://www.docker.com/) (for local development)
- AWS account (for cloud deployment)

---

### 🐳 Run Locally with Docker

The easiest way to run the full stack locally:

```bash
git clone https://github.com/Pratik1603/GO-STREAM-AWS-.git
cd GO-STREAM-AWS-
```

**Start the backend:**
```bash
cd server
docker build -t gostream-server .
docker run -p 8080:8080 gostream-server
```

**Start the frontend:**
```bash
cd client
npm install
npm run dev
```

The client will be available at `http://localhost:3000` and the server at `http://localhost:8080`.

---

### 🛠️ Run Without Docker

```bash
# Backend
cd server
go mod tidy
go run main.go

# Frontend (in a new terminal)
cd client
npm install
npm run dev
```

---

## ☁️ AWS Serverless Deployment

The backend is deployed as a **serverless function on AWS Lambda**, triggered via **API Gateway**.

| Component         | Service              |
|-------------------|----------------------|
| **Backend**       | AWS Lambda (Go)      |
| **API Layer**     | AWS API Gateway      |
| **Media Storage** | AWS S3               |
| **Frontend**      | Vercel               |

### Deploy to AWS Lambda

```bash
cd server

# Build Go binary for Linux (Lambda runtime)
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Zip for Lambda upload
zip function.zip bootstrap

# Deploy via AWS CLI
aws lambda update-function-code \
  --function-name gostream \
  --zip-file fileb://function.zip
```

---

## 🛠️ Tech Stack

| Layer       | Technology              |
|-------------|-------------------------|
| Backend     | Go (Golang)             |
| Serverless  | AWS Lambda              |
| API         | AWS API Gateway         |
| Storage     | AWS S3                  |
| Frontend    | JavaScript / React      |
| Local Dev   | Docker                  |
| Hosting     | Vercel (Frontend)       |

---

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

<div align="center">
  Made with ❤️ by <a href="https://github.com/Pratik1603">Pratik1603</a>
</div>
