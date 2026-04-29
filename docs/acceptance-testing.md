# WorkPal Acceptance Testing Guide

This guide is for manually validating the current project after local startup.

## Before you start

Confirm all three conditions:

1. Docker dependencies are running
2. Backend health check returns `200`
3. Frontend is reachable at `http://localhost:3000`

## Acceptance accounts

| Role | Username | Password | Suggested usage |
|---|---|---|---|
| Admin | `admin` | `admin123` | full workspace walkthrough |
| Employee | `emma.chen` | `workpal123` | direct chat and group member |
| Employee | `liam.wang` | `workpal123` | engineering and directory checks |
| Employee | `sofia.zhao` | `workpal123` | design and release-readiness checks |

## Recommended manual walkthrough

### 1. Login page

Validate:

- seeded accounts are visible
- language can switch between `English / 简体中文`
- theme can switch between light and dark

### 2. Overview

Validate:

- overview loads without API errors
- metric cards render
- clicking cards or module buttons jumps to the correct workspace section

### 3. Preferences

Open the preferences drawer and validate:

- language switch
- theme switch
- message sound toggle
- comfortable and compact density switch

### 4. Directory

Validate:

- seeded employees are visible
- department filter works
- fuzzy search works for title, phone, employee number, and department

Suggested checks:

- choose `Engineering`
- search `Platform Engineer`
- confirm `liam.wang` remains visible

### 5. Chat

Validate direct chat:

- create a direct chat with `emma.chen`
- send a message

Validate group chat:

- create a group with `emma.chen` and `liam.wang`
- send a group message
- edit the group announcement
- upload a group file

### 6. Tasks

Validate:

- create a task
- move it across columns
- share it
- delete it

### 7. Schedule

Validate:

- create an event
- share it
- delete it

### 8. Files

Validate:

- upload a personal file
- open it
- share it
- delete it

## Automated smoke check

With backend and frontend already running:

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

The smoke test currently covers:

- health and metrics endpoints
- seeded login API
- direct and group chat API flows
- group announcement and group file API flows
- frontend login
- overview navigation
- directory filter and search
- task creation
- schedule creation
- file upload
- direct chat creation and send
