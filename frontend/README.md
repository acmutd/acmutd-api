# ACM UTD API Portal Frontend (Vite + React + TypeScript)

## Stack

- Vite + React 18 + TypeScript
- React Router
- Firebase Auth (Google sign-in)
- Tailwind utility classes

## Architecture

Firebase config is **not** stored in the frontend. The Go backend injects it at request time by replacing the `__FIREBASE_CONFIG__` placeholder in `index.html` (see `internal/server/handlers/dashboard_pages.go`). The frontend reads it from `window.__FIREBASE_CONFIG__` in `src/lib/firebase.ts`.

All API calls attach a Firebase ID token as `Authorization: Bearer <token>`. The backend verifies it via the Firebase Admin SDK.

### Key files

| Path | Purpose |
|------|---------|
| `src/lib/firebase.ts` | Firebase app + auth init |
| `src/lib/api.ts` | Backend API calls |
| `src/lib/apiClient.ts` | HTTP client with auth header |
| `src/state/AuthContext.tsx` | Google sign-in, auth state |
| `src/types/models.ts` | Shared data models |
| `src/pages/` | Route-level page components |
| `src/components/` | Reusable UI components |

## Routes

- `/` — Landing / sign-in
- `/dashboard` — User dashboard
- `/admin` — Admin dashboard (requires admin role)

## Development

The Go backend must be running since it injects the Firebase config into the HTML. Visit `http://localhost:8080` instead of the Vite dev server URL.

```bash
# In the repo root
go run ./cmd/api/main.go
```

If you need Vite HMR during development, start both servers and proxy the Go backend in `vite.config.js`.

## Production build

```bash
npm install
npm run build
```

Output goes to `frontend/dist`. The Go backend serves it and injects the Firebase config at request time — **do not** embed Firebase credentials in the frontend build.
