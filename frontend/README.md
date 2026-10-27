# Måltiden frontend

Vue 3 + TypeScript + Vite single-page app. Setup, checks and project layout
are in the [root README](../README.md).

```sh
npm install
npm run dev          # against the Go backend on http://localhost:8080
npm run dev:mock     # mocked API (src/mocks), no backend needed
npm run type-check
npm run lint         # eslint --fix
npm test             # vitest
npm run build
```

`VITE_API_URL` and `VITE_USE_REAL_API` come from `.env.development` and `.env.mock`.
Production builds use same-origin API calls, and the Go server serves `dist/`.
