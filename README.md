# Internal Leads

Aplikasi internal tim sales ISP korporat: input lead, catat follow-up berulang, handoff ke Odoo saat lead siap survey.

Data model, API contract, dan aturan bisnis lengkap ada di [ARCHITECTURE.md](ARCHITECTURE.md).

## Directory Structure

- [backend/](backend/): Server Go (Fiber) — API, migration (`golang-migrate`), model, seeder wilayah
- [frontend/](frontend/): Aplikasi React + Vite + TypeScript (belum diisi)

## Tech Stack

- Go + Fiber
- GORM + PostgreSQL
- golang-migrate
- React + Vite + TypeScript
- Docker multi-stage (frontend di-embed ke binary Go)

## Git Guidelines

Please follow these guidelines to ensure that all commits are consistent and informative.

### Commit Message Structure

Commit messages should follow this structure:
- **Tag**: The type of change (e.g., `[FIX]`, `[IMP]`, `[ADD]`, etc.).
- **Scope**: The technical name of the module/package/service being modified.
- **Short Description**: A concise summary of the change (ideally < 50 characters).
- **Blank Line**: Always leave one empty line between the header and the body.
- **Full Description**: Explain the reasoning behind the change as **bullet points**, focusing on the *why* rather than the *what*. Include any relevant task numbers, PR numbers, or references as separate points.

Example:

```
[FIX] scope_name: short description

* Detailed explanation of the change, why it was needed, and any technical decisions made.
* References: task-123
* Fixes #456
```

### Tags

Use one of the following tags to prefix your commit message:
- **[FIX]** for bug fixes: mostly used in stable version but also valid if you are fixing a recent bug in development version
- **[IMP]** for improvements: most of the changes done in development version are incremental improvements not related to another tag
- **[ADD]** for adding new modules/features
- **[REF]** for refactoring: when a feature is heavily rewritten
- **[REM]** for removing resources: removing dead code, removing views, removing modules, …
- **[REV]** for reverting commits: if a commit causes issues or is not wanted reverting it is done using this tag
- **[MOV]** for moving files: use git move and do not change content of moved file otherwise Git may loose track and history of the file; also used when moving code from one file to another
- **[REL]** for release commits: new major or minor stable versions
- **[MERGE]** for merge commits: used in forward port of bug fixes but also as main commit for feature involving several separated commits
- **[I18N]** for changes in translation files

### Commit Message Header

The header should be a meaningful and concise summary of the change. It should make sense when combined with the phrase "if applied, this commit will...". Try to limit the header length to around 50 characters for readability.

### Full Description

Always leave a blank line after the header, then write the full description as bullet points. Focus on explaining the *why* behind the change. If there were any technical choices involved, explain those as well. Avoid making commits that affect multiple modules at once; try to split changes into separate commits for each scope.

Examples of proper commit messages:

```
[REF] models: use `parent_path` to implement parent_store

* Replaces the former modified preorder tree traversal (MPTT) with the fields `parent_left`/`parent_right`.
* Improves read performance on hierarchical queries.
```

```
[FIX] auth: remove hardcoded token expiry

* Token expiry is now read from config instead of a magic number.
* Fixes #12345
```

**Take the time to write clear and understandable commit messages, as they are crucial for maintaining a clean and traceable project history.**

## Setup Instructions

1. **Clone the Repository**

   Remote belum dikonfigurasi untuk repo ini. Setelah remote tersedia: `git clone <url>`.

2. **Configure Git**

   Make sure both `user.email` and `user.name` are defined in your git config:
   ```bash
   git config --global user.name "Your Full Name"
   git config --global user.email "your.email@example.com"
   ```

3. **Backend — Install & Configure**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # edit .env sesuai kebutuhan (DB, JWT secret, dll.)
   ```

4. **Backend — Migrate & Run**
   ```bash
   migrate -path migrations -database "$DATABASE_URL" up
   go run ./cmd/api      # → :3000
   ```

5. **Backend — Seed data wilayah** (sekali, setelah migrate; baca CSV dari `backend/seeds/`)
   ```bash
   go run ./cmd/seed
   ```

6. **Frontend**

   Belum ada isinya — instruksi setup akan ditambahkan saat fase frontend dimulai.

7. **Create a New Branch**

   Before making any changes, create a new branch based on the task you are working on:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Deployment

Single Docker image (frontend embedded into the Go binary), deployed as a
Portainer Git stack behind an existing Nginx Proxy Manager container.

### Test locally first

```bash
cp .env.example .env      # fill in real-looking values, any values work locally
docker compose -f docker-compose.yml -f docker-compose.local.yml up --build -d
```

App is reachable at `http://localhost:3000`. First run only, apply the
schema and reference data:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml exec leadsales-app ./migrate-up.sh
docker compose -f docker-compose.yml -f docker-compose.local.yml exec leadsales-app ./seed
```

Log in with the bootstrap account (`admin@leadsales.local` /
`ChangeMe123!`) and change the password immediately.

### Deploy to the VM

1. In Nginx Proxy Manager, note the Docker network name it runs on
   (`docker network ls` on the VM).
2. In Portainer: **Stacks → Add stack → Repository**, point it at this
   repo, compose path `docker-compose.yml`. Fill in the stack's
   **Environment variables** using `.env.example` as the checklist, and set
   `PROXY_NETWORK_NAME` to the value from step 1.
3. Deploy. Postgres becomes healthy, then the app starts. The app connects
   to the database, not to any table, so it stays up even before the schema
   exists - only requests that touch the database return errors until
   migrations are applied.
4. **First deploy only** - open the `leadsales-app` container's Console in
   Portainer and run:
   ```sh
   ./migrate-up.sh
   ./seed
   ```
   Then log in with the bootstrap account above and change its password.
5. In Nginx Proxy Manager, add a proxy host pointing at `leadsales-app`,
   port `3000`, scheme `http` (NPM terminates TLS; the app trusts
   `APP_ENV=production` alone to mark its refresh-token cookie `Secure`, so
   no extra proxy header configuration is needed).

### Subsequent deploys

Redeploy the stack from Portainer (pulls the latest commit, rebuilds the
image, recreates the container - data is untouched, it lives in the
`pgdata` volume). Only re-run `./migrate-up.sh` when a new migration was
actually added, and review it before running.
