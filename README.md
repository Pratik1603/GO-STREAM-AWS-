# 🎬 GO STREAM

A Netflix-style video streaming platform built with **Go** (backend) and deployed on **AWS + Vercel**. Stream movies and TV shows with AI-powered recommendations, S3-backed video delivery, and presigned URL support.

---

## 📸 Screenshots

### 🏠 Home — AI Recommended
![Home Page](./screenshot-home.png)
> AI-curated movie recommendations on the landing page with hero banner, Play & More Info actions.

---

### 🎥 Movie Detail Page
![Movie Detail](./screenshot-detail.png)
> Full movie detail view with genre tags, year, description, and Admin Controls (Edit / Delete) for authorized users.

---

### ➕ Add New Content (Admin)
![Add Content](./screenshot-add.png)
> Admin panel to upload new Movies, TV Series, or Episodes — with fields for title, description, release year, director, and genres.

---

## 🚀 Features

- 🎞️ **Stream movies** via AWS S3 presigned URLs (direct-to-client, no server bandwidth used)
- 🤖 **AI Recommendations** powered by Claude
- 🔐 **Auth middleware** with role-based access (admin vs user)
- 📂 **S3 file storage** for video and thumbnail assets
- 🧩 **Content types**: Movies, TV Series, Episodes
- 🔎 **Search** across titles
- 🔔 Notifications support

---

## 🛠️ Tech Stack

| Layer | Tech |
|---|---|
| Backend | Go + Gin |
| Hosting | AWS Lambda / EC2 |
| Storage | AWS S3 |
| Frontend | React (Vite) |
| Deployment | Vercel |
| Auth | JWT middleware |

---

## ⚙️ API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/movies/:id/stream` | Stream or redirect to S3 presigned URL |
| `POST` | `/upload?key=` | Upload file to S3 |
| `GET` | `/file?key=` | Fetch file from S3 |
| `GET` | `/presign?key=&method=` | Generate presigned URL (15 min) |

---

## 🏗️ Project Structure

```
├── main.go               # Lambda entry point
├── handlers/
│   └── movie.go          # StreamMovie, GetMovie handlers
├── services/
│   └── movie_service.go  # Business logic, S3 operations
├── middleware/
│   └── auth.go           # JWT auth, userID injection
└── models/
    └── movie.go          # Movie struct
```

---

## 🔧 Setup & Deployment

### 1. Clone & install
```bash
git clone https://github.com/your-username/go-stream
cd go-stream
go mod tidy
```

### 2. Environment variables
```env
AWS_REGION=us-east-1
AWS_BUCKET_NAME=your-bucket
JWT_SECRET=your-secret
```

### 3. Build for Lambda
```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip lambda.zip bootstrap
```

### 4. Frontend (React + Vercel)
```bash
cd frontend
npm install
npm run build
```

Add `vercel.json` to fix SPA routing on refresh:
```json
{
  "rewrites": [{ "source": "/(.*)", "destination": "/index.html" }]
}
```

---

## 📝 License

MIT
