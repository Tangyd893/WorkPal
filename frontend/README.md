# WorkPal Frontend

This directory contains the React + Vite frontend for WorkPal.

## What the frontend expects

The frontend talks only to the API gateway at `http://localhost:8080`.

Proxy rules in `vite.config.ts`:

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

That means the frontend stays stable even though the backend is split into multiple domain services behind the gateway.

## Start the frontend

The frontend uses npm as the only package manager. Keep `package-lock.json` as the dependency lock file and install with `npm ci`.

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

Open:

```text
http://localhost:3000
```

## Workspace modules

After login, the app routes into a workspace shell with these modules:

- Overview
- Chat
- Tasks
- Schedule
- Files
- Directory

## User-facing capabilities

- seeded acceptance accounts shown on the login page
- language switch: `English / 简体中文`
- light and dark theme
- message sound toggle
- density toggle
- direct chat and group chat
- group announcement and group files
- directory search by name, phone, title, employee number, and department
- task create, update, share, and delete
- schedule create, share, and delete
- file upload, share, and delete

## Source layout

- `src/pages`: route-level pages such as `LoginPage`, `WorkspacePage`, and `ChatPage`
- `src/components`: workspace and chat UI components
- `src/api`: API wrappers and response unwrapping
- `src/hooks`: auth, preferences, and chat controller state
- `src/types`: shared TypeScript models
- `src/data`: seeded login account display data
- `src/styles`: global styles
- `src/test`: small React/Vitest render helpers used by component tests

## Backend-backed vs display-only data

### Backend-backed

- login
- current user
- users and departments
- direct and group chat
- message history and search
- group announcement
- group files
- personal files
- tasks
- schedule

### Display-only right now

- overview summary composition

## Non-functional baseline

- package management is npm-only
- ESLint is configured for `src`, `vite.config.ts`, and `vitest.config.ts`
- Vitest covers API response unwrapping, auth/chat stores, preference persistence, i18n, login form semantics, settings drawer semantics, and file upload panel behavior
- error banners use `role="alert"`; loading and informational banners use `role="status"`
- modal surfaces use dialog semantics, and the settings drawer closes on Escape
- segmented controls and selectable cards expose their selected state with ARIA attributes
- global styles include keyboard `focus-visible` treatment and reduced-motion support

## Tests

```powershell
cd frontend
npm run lint
npm test
npm run build
```

End-to-end smoke:

```powershell
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```
