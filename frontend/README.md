# ACM UTD API Portal Frontend (Vite + React + TypeScript)

This frontend is a pure mocked skeleton of the ACM UTD API Portal.

## Routes

- `/` Landing/auth stub
- `/dashboard` User dashboard
- `/admin` Admin dashboard (mock admin-only)

## Stack

- Vite
- React 18
- TypeScript
- React Router
- Tailwind utility classes

## Mock architecture

- Mock Firestore-shaped models: `src/types/models.ts`
- In-memory mock documents: `src/lib/mockData.ts`
- Async no-op/mock API layer: `src/lib/api.ts`

All API integration points return promises and include artificial delay for realistic UI behavior.

## Development

```bash
npm install
npm run dev
```

## Production build

```bash
npm run build
```

Build output is generated in `frontend/dist`.
