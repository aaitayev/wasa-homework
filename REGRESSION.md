# Regression Test Report

## Overview
This document outlines the end-to-end regression scenario for the WASA project and documents the verification results.

## Regression Scenario (Step-by-Step)

### 1. Authentication
- [x] **Alice logs in**: `POST /session` -> `201 Created`
- [x] **Bob logs in**: `POST /session` -> `201 Created`
- [x] **Persistence**: Verify tokens are valid after server restart.

### 2. Messaging & Conversations
- [x] **Direct Message**: Alice sends "Hello Bob!" to Bob.
- [x] **Inbox Update**: Bob checks `GET /conversations` and sees Alice's message.
- [x] **Replying**: Bob replies "Hi Alice!".
- [x] **Conversation View**: Alice checks `GET /conversations/{id}` and sees both messages.
- [x] **Ordering**: Messages are ordered by `created_at` ASC.

### 3. Advanced Messaging Features
- [x] **Commenting**: Alice comments on Bob's message.
- [x] **Soft-deletion**: Bob deletes his message.
- [x] **Conflict Handling (409)**: Alice tries to comment or forward the deleted message -> **Rejected with 409**.
- [x] **Inbox Refresh**: Bob's message shows as deleted in the conversation view.

### 4. Group Management
- [x] **Group Creation**: Alice creates group "Project X" with Bob.
- [x] **Renaming**: Alice renames group to "Project Y".
- [x] **Leaving**: Bob leaves the group.
- [x] **Group Photos**: Alice uploads and retrieves a group photo.

### 5. Media & Profiles
- [x] **Profile Photo**: Alice uploads a PNG photo.
- [x] **Content-Type**: Verify `GET /me/photo` returns `image/png`.
- [x] **Size Validation**: Uploading >5MB fails with `413`.
- [x] **Type Validation**: Uploading non-image fails with `415`.

### 6. Infrastructure & Persistence
- [x] **Docker Build**: `docker compose up --build` works from a clean state.
- [x] **API Proxy**: Frontend connects to `/api` which proxies to `backend:3000`.
- [x] **Volume Persistence**: Data survives `docker compose restart`.
- [x] **Hard Reset**: `docker compose down -v` wipes the database.

## Automated Verification
Run the following script to verify the above scenario:
```bash
./scripts/regression_test.sh
```
*Note: Ensure the backend is running at localhost:3000 or set `API_URL` environment variable.*

## Manual Verification Proof
- **Liveness Check**: `curl -f http://localhost:3000/liveness` -> `OK`
- **DB Check**: `sqlite3 data/wasa.db "SELECT * FROM users;"`
