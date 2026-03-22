# Job Pipeline with SSE

Playing around with bun, drizzle, sqlite, and server-sent events. Created this to figure out how to do real-time job status updates without the headache of websockets.

## What's this?

A simple job processing system where you can:
- Create jobs with some payload
- Process them (which simulates some work with delays)
- Watch their status update in real-time via Server-Sent Events (SSE)
- Run multiple UIs to compare React vs Svelte behavior

The backend is a Hono app running on bun, using Drizzle ORM with SQLite. Two frontend implementations exist - one in React and one in Svelte - mostly to see if the weird React-specific bugs I was hitting would show up in Svelte too.

Spoiler: they did, so it wasn't a React issue after all.

## Project Structure

```
.
├── src/                    # Backend (Bun + Hono + Drizzle)
│   ├── app.ts             # Main Hono app with routes
│   ├── db.ts             # Database connection and storage
│   ├── db.schema.ts      # Drizzle schema for jobs table
│   ├── processor.ts      # Job processing logic (with simulated delays)
│   ├── logger.ts         # Request logging
│   └── index.ts          # Entry point
│
├── src/ui/               # React Frontend
│   ├── src/
│   │   ├── api/          # API client
│   │   ├── components/   # React components
│   │   └── App.tsx
│   └── vite.config.ts
│
└── src/ui-svelte/        # Svelte Frontend (experimental)
    ├── src/
    │   ├── lib/          # Components and API
    │   └── App.svelte
    └── vite.config.ts
```

## Setup

```bash
# Install backend deps
bun install

# Install React UI deps
cd src/ui && bun install

# Install Svelte UI deps
cd src/ui-svelte && bun install
```

## Running

```bash
# Start backend (port 3000)
bun run src/index.ts

# Start React UI (port 5173)
cd src/ui && npm run dev

# Start Svelte UI (port 5174)
cd src/ui-svelte && npm run dev
```

Or just run the backend and visit both UIs - they're both pointing to `localhost:3000` for the API.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api` | Health check |
| POST | `/api/create-job` | Create a new job |
| POST | `/api/process/:jobId` | Start processing a job |
| GET | `/api/jobs/status?jobIds=...` | SSE stream for job updates |

### Job Status Flow

```
idle → queued → processing → completed
                    ↓
                failed (on error)
```

The processor simulates work with 3-second delays between each status transition. So idle → completed takes about 9 seconds.

## SSE Stuff

The `/api/jobs/status` endpoint streams updates in SSE format:

```
event: job-status
data: {"id":"...","status":"completed","result":"..."}

event: done
data: all jobs completed
```

There's probably some edge cases with the SSE handling that I haven't found yet. If you see weird stuff, check the browser console and backend logs.

## What I Learned

- SSE is way simpler than websockets for one-way updates
- EventSource can be finicky about connection state
- Multiple SSE connections to the same endpoint can cause race conditions
- React's functional state updates vs Svelte's reactive assignment behave slightly different when updates come in rapid succession
- `neverthrow` Result type is nice but adds verbosity

## TODO / Known Issues

- [ ] Input validation on job creation
- [ ] Delete job endpoint
- [ ] Proper error handling in the SSE stream
- [ ] Maybe add a timeout for jobs that get stuck
- [ ] Tests (I know, I know)
- [ ] This README could use more detail

## Tech Stack

- **Backend**: Bun, Hono, Drizzle ORM, SQLite, neverthrow
- **React UI**: React 19, Tailwind, Vite
- **Svelte UI**: Svelte 5, Tailwind, Vite
- **Database**: SQLite (because it's there, not because it's the best choice)

Built with bun init because I wanted to try it out. The project works but I wouldn't call it production-ready.
