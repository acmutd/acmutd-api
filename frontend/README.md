# Frontend Dashboard (Vite + Tailwind)

This folder contains the frontend pages served by the Go API.

## Pages

- Dashboard entry: `src/dashboard/dashboard.html` -> served at `/dashboard` and `/admin`
- Docs landing page: `src/swagger/swagger.html` -> served at `/` by the backend

The dashboard page uses Tailwind stylesheet at `src/dashboard/tailwind.css`.

## Development

```bash
npm install
npm run dev
```

## Production Build

```bash
npm install
npm run build
```

Build output is written to `frontend/dist`.

The backend expects built files in:

- `frontend/dist/src/dashboard/dashboard.html`
- `frontend/dist/assets/*`
