# Notree

## Quick start

**Clone the repository:**

```bash
git clone https://github.com/KubantsevAS/notree.git
cd notree
```

**Set up environment variables:**

```bash
cp .env.example .env
```

**Launch DB:**

```bash
docker compose up -d
```

**Launch DB migrations (with Taskfile):**

```bash
cd backend
task migrate
```

**If Taskfile not installed:**

```bash
cd backend
migrate -path ./migrations -database "postgres://YOUR_POSTGRES_USER:YOUR_POSTGRES_PASSWORD@localhost:5432/YOUR_POSTGRES_DB?sslmode=disable" up
```

## Drop DB && Rebuild App

```bash
docker compose down -v
docker compose up --build -d
```

## Install Taskfile

```bash
sudo snap install task --classic
```
