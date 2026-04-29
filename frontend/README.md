# WorkPal Frontend

This is the React + Vite frontend for WorkPal.

If you want to boot the whole project, read the repo root [README.md](../README.md) first. This file focuses on frontend structure and learning points.

## Frontend stack

- React 18
- Vite
- TypeScript
- Zustand
- Axios
- Playwright for smoke testing

## Start the frontend

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

Open:

```text
http://localhost:3000
```

## Proxy behavior

Defined in [vite.config.ts](vite.config.ts):

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

That means the frontend expects the backend to be running on `http://localhost:8080`.

## Main pages and modules

After login, the app routes into a workspace shell:

- `Overview / 总览`
- `Chat / 沟通`
- `Tasks / 任务`
- `Schedule / 日程`
- `Files / 文件`
- `Directory / 通讯录`

## User-facing capabilities

- seeded acceptance accounts shown on the login page
- language switch: `English / 简体中文`
- light and dark theme
- message sound toggle
- compact density toggle
- direct chat and group chat
- group announcement and group files
- directory search by name, phone, title, employee number, and department
- task CRUD and share actions
- schedule CRUD and share actions
- file upload, open, share, and delete actions

## Source layout

Key frontend folders:

- `src/pages`: route-level pages such as `LoginPage` and `WorkspacePage`
- `src/components`: module UIs for workspace and chat
- `src/api`: backend request wrappers
- `src/hooks`: Zustand-backed auth, preferences, and chat state logic
- `src/types`: shared frontend types
- `src/data`: seeded display data used by the overview and knowledge cards
- `src/utils`: storage, clipboard, and notification helpers
- `src/styles`: global CSS

## Backend-backed vs seeded UI data

### Backend-backed

- login
- current user
- users and departments
- direct and group chat
- message history and search
- group announcement
- group files
- personal file uploads
- tasks
- schedule

### Seeded or display-only data

- overview summaries
- seeded knowledge cards in the files module

So the workspace shell mixes real backend data with a few curated demo cards to make the product surface more complete during acceptance.

## State management notes

- auth state lives in Zustand and persists in local storage
- UI preferences also persist in local storage
- chat runtime state is coordinated through `useChatController`
- most module actions are routed through `workpalApi`

## Tests

Unit and build checks:

```powershell
cd frontend
npm test
npm run build
```

End-to-end smoke:

```powershell
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```
