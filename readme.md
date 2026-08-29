# Notree

## Quick start

1. **Clone the repository:**

   ```bash
   git clone https://github.com/KubantsevAS/notree.git
   cd notree
   ```

2. **Set up environment variables:**

   ```bash
   cp .env.example .env
   ```

3. **Launch DB:**

   ```bash
   docker compose up -d
   ```

4. **Launch DB migrations**

   * Option A: with Taskfile:

    ```bash
    cd backend
    task migrate
    ```

   * Option B: if Taskfile not installed:

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
