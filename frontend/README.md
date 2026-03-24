# Mytoken Svelte Frontend

This is the new Svelte-based frontend for the mytoken server, replacing the legacy Mustache templates.

## Technology Stack

- **SvelteKit 2** - Modern web framework with SSG support
- **TypeScript** - Type-safe JavaScript
- **Bootstrap 5** - CSS framework
- **Vite** - Fast build tool

## Development

### Prerequisites

- Node.js 18+ (22+ recommended)
- npm 9+

### Install Dependencies

```bash
cd frontend
npm install
```

### Development Server

```bash
npm run dev
```

The dev server runs at http://localhost:5173 and proxies API requests to http://localhost:8000 (the Go backend).

### Type Checking

```bash
npm run check
```

### Build for Production

```bash
npm run build
```

The built files are output to `frontend/build/`.

> **Note:** The `internal/server/spa/dist/` directory is gitignored. The CI pipeline automatically builds the frontend and includes it in the release artifacts. For local development, use `./scripts/build-frontend.sh` to build and copy the frontend to the correct location.

## Go Backend Integration

### Building and Embedding

Use the provided script to build the frontend and copy it to the Go embed directory:

```bash
./scripts/build-frontend.sh
```

This will:
1. Build the Svelte frontend
2. Copy the built files to `internal/server/spa/dist/`
3. The files will be embedded in the Go binary during compilation

### Enabling the SPA

Add the following to your `config.yaml`:

```yaml
features:
  web_interface:
    enabled: true
    use_spa: true
```

When `use_spa` is `true` and the SPA distribution is available, the server will serve the Svelte SPA instead of the Mustache templates.

### Fallback Behavior

- If `use_spa` is `false` or the SPA files are not available, the server falls back to the original Mustache templates
- API routes (`/api/*`) are always handled by the Go backend
- The SPA handles client-side routing for all other paths

## Project Structure

```
frontend/
├── src/
│   ├── app.html           # HTML shell
│   ├── app.d.ts           # TypeScript declarations
│   ├── routes/            # SvelteKit routes (pages)
│   │   ├── +layout.svelte # Main layout
│   │   ├── +page.svelte   # Home page
│   │   ├── settings/      # Settings pages
│   │   ├── consent/       # Consent page
│   │   └── ...
│   └── lib/               # Shared code
│       ├── api/           # API client
│       ├── components/    # Svelte components
│       ├── stores/        # Svelte stores
│       ├── types/         # TypeScript types
│       └── utils/         # Utility functions
├── static/                # Static assets
├── package.json
├── svelte.config.js
├── vite.config.ts
└── tsconfig.json
```

## Components

### Atomic Components
- `TagPill` - Display/remove tags
- `CopyButton` - Copy to clipboard button
- `LoadingSpinner` - Loading indicator
- `CollapsibleSection` - Expandable sections
- `ErrorModal` - Error display modal
- `ToastContainer` - Toast notifications

### Composite Components
- `CapabilityTree` - Capability selection tree
- `RestrictionsEditor` - Restriction clause editor
- `RotationSettings` - Token rotation settings

### Token Components
- `CreateMytoken` - Create new mytoken form
- `TokenInfo` - Token introspection display
- `TokenList` - List of user's tokens
- `CreateAccessToken` - Get access token form
- `TransferCode` - Create/redeem transfer codes

## API Integration

The frontend communicates with the Go backend via the same REST API used by the original frontend. The API client in `src/lib/api/client.ts` provides typed methods for all endpoints:

- Token operations (create, introspect, revoke)
- User settings (grants, SSH keys, email)
- Notifications (calendars, subscriptions)
- Transfer codes

## Migration from Mustache

The Svelte frontend is designed as a drop-in replacement:

1. Same Bootstrap styling (upgraded to v5)
2. Same API endpoints
3. Same authentication flow
4. Configurable via `use_spa` option

Both frontends can coexist - set `use_spa: false` to use the original Mustache templates.
