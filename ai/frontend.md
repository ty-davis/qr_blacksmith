# Frontend Plan — Vue 3 + TypeScript + Tailwind CSS

## Project Bootstrap

```bash
npm create vue@latest frontend
# Options: TypeScript ✓, Vue Router ✓, Pinia ✓, ESLint ✓, Prettier ✓
cd frontend
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
npm install axios @vueuse/core chart.js vue-chartjs
```

---

## Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts            -- axios instance with base URL + auth header + token refresh
│   │   ├── auth.ts              -- register, login, logout, refresh, me
│   │   ├── batches.ts           -- batch API calls
│   │   ├── qrcodes.ts           -- QR code API calls
│   │   └── analytics.ts         -- analytics/dashboard API calls
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppSidebar.vue
│   │   │   └── AppHeader.vue
│   │   ├── batches/
│   │   │   ├── BatchCard.vue    -- summary card (name, code count, scan count)
│   │   │   └── BatchForm.vue    -- create/edit form
│   │   ├── qrcodes/
│   │   │   ├── QRCodeCard.vue   -- image + label + scan count + redirect URL
│   │   │   ├── QRCodeGrid.vue   -- paginated grid of QRCodeCards
│   │   │   └── GenerateModal.vue -- modal to generate N codes for a batch
│   │   ├── analytics/
│   │   │   ├── ScanLineChart.vue  -- scans over time (Chart.js)
│   │   │   ├── DevicePieChart.vue -- device type breakdown
│   │   │   ├── CountryTable.vue   -- top countries table
│   │   │   └── StatCard.vue       -- single number stat (reusable)
│   │   └── shared/
│   │       ├── ConfirmModal.vue
│   │       ├── Pagination.vue
│   │       └── EmptyState.vue
│   ├── stores/
│   │   ├── auth.ts              -- Pinia store: user, accessToken, login/logout actions
│   │   ├── batches.ts           -- Pinia store for batches
│   │   └── qrcodes.ts           -- Pinia store for QR codes
│   ├── views/
│   │   ├── LoginView.vue        -- /login
│   │   ├── RegisterView.vue     -- /register
│   │   ├── DashboardView.vue    -- /
│   │   ├── BatchListView.vue    -- /batches
│   │   ├── BatchDetailView.vue  -- /batches/:id
│   │   └── QRCodeDetailView.vue -- /batches/:id/qrcodes/:qrId
│   ├── router/
│   │   └── index.ts             -- routes + auth guard
│   ├── types/
│   │   └── index.ts             -- shared TypeScript interfaces
│   ├── App.vue
│   └── main.ts
├── tailwind.config.ts
├── vite.config.ts
└── package.json
```

---

## TypeScript Types

```typescript
// src/types/index.ts

export interface User {
  id: string
  email: string
  created_at: string
}

export interface Batch {
  id: string
  name: string
  description: string | null
  redirect_url: string
  qr_code_count: number
  total_scans: number
  created_at: string
  updated_at: string
}

export interface QRCode {
  id: string
  hash: string
  batch_id: string
  redirect_url: string | null     // null = inherits from batch
  effective_url: string           // always resolved
  label: string | null
  scan_count: number
  last_scanned_at: string | null
  qr_image_url: string
  created_at: string
}

export interface ScanAnalytics {
  total_scans: number
  unique_codes_scanned?: number
  scans_by_day: { date: string; count: number }[]
  top_countries: { country: string; country_code: string; count: number }[]
  device_breakdown: { device_type: string; count: number }[]
}

export interface OverviewStats {
  total_batches: number
  total_qr_codes: number
  total_scans: number
  scans_today: number
  scans_this_week: number
  most_scanned_code: QRCode | null
  recent_scans: RecentScan[]
}

export interface RecentScan {
  qr_code_id: string
  qr_code_label: string | null
  batch_name: string
  scanned_at: string
  city: string | null
  country: string | null
  device_type: string
}
```

---

## Router & Auth Guard

```typescript
// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/login',    component: () => import('@/views/LoginView.vue'),    meta: { public: true } },
  { path: '/register', component: () => import('@/views/RegisterView.vue'), meta: { public: true } },
  { path: '/',         component: () => import('@/views/DashboardView.vue') },
  { path: '/batches',  component: () => import('@/views/BatchListView.vue') },
  { path: '/batches/:id', component: () => import('@/views/BatchDetailView.vue') },
  { path: '/batches/:id/qrcodes/:qrId', component: () => import('@/views/QRCodeDetailView.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.public && auth.isLoggedIn) {
    return { path: '/' }
  }
})

export default router
```

---

## Views

### `/login` — Login
- Email + password form
- On success: stores access token, redirects to `/` (or `?redirect` param)
- Link to `/register`

### `/register` — Register
- Email + password + confirm password
- On success: same as login flow

### `/` — Dashboard
- Four `StatCard` components: total codes, total scans, scans today, scans this week
- `ScanLineChart` — last 30 days across all codes
- Recent scans table (last 10 events)
- Quick link to most-scanned code

### `/batches` — Batch List
- Grid of `BatchCard` components
- "New Batch" button → opens `BatchForm` inline or in a modal
- Each card links to `/batches/:id`

### `/batches/:id` — Batch Detail
- Batch name, description, redirect URL (editable inline)
- **Bulk redirect update section** — input for new URL + toggle for
  "Override individual QR code URLs too" — calls `PUT /api/batches/:id/redirect`
- "Generate QR Codes" button → `GenerateModal`
- `ScanLineChart` for this batch
- `CountryTable` + `DevicePieChart` for this batch
- `QRCodeGrid` with all codes in the batch

### `/batches/:id/qrcodes/:qrId` — QR Code Detail
- Large QR code image with download button (PNG)
- Label (editable inline)
- Personal redirect URL (editable, clearable to revert to batch)
- `ScanLineChart` for this code
- `CountryTable` + `DevicePieChart` for this code
- Scan history table (paginated)

---

## Auth Store (`stores/auth.ts`)

```typescript
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    accessToken: null as string | null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.accessToken,
  },
  actions: {
    async login(email: string, password: string) {
      const { access_token, user } = await authApi.login(email, password)
      this.accessToken = access_token
      this.user = user
    },
    async register(email: string, password: string) {
      const { access_token, user } = await authApi.register(email, password)
      this.accessToken = access_token
      this.user = user
    },
    async refresh() {
      // Called automatically by the axios interceptor on 401
      const { access_token } = await authApi.refresh()
      this.accessToken = access_token
    },
    async logout() {
      await authApi.logout()
      this.accessToken = null
      this.user = null
    },
  },
})
```

---

## API Client

```typescript
// src/api/client.ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,  // send the HttpOnly refresh token cookie
})

// Attach access token to every request
client.interceptors.request.use(config => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

// On 401, attempt one silent token refresh then retry
let isRefreshing = false
client.interceptors.response.use(
  res => res,
  async err => {
    const original = err.config
    if (err.response?.status === 401 && !original._retry) {
      original._retry = true
      if (!isRefreshing) {
        isRefreshing = true
        try {
          await useAuthStore().refresh()
        } finally {
          isRefreshing = false
        }
      }
      return client(original)
    }
    const message = err.response?.data?.error ?? 'An unexpected error occurred'
    return Promise.reject(new Error(message))
  }
)

export default client
```

---

## Environment Variables

```env
# frontend/.env
VITE_API_BASE_URL=http://localhost:8080
```

```env
# frontend/.env.production
VITE_API_BASE_URL=https://api.yourdomain.com
```

---

## Key UX Behaviours

| Interaction | Behaviour |
|---|---|
| Login / register | Redirect to `/` on success; show field-level validation errors |
| Token expiry | Axios interceptor silently refreshes; user never notices |
| Logout | Calls `POST /api/auth/logout`, clears store, redirects to `/login` |
| Generate QR codes | Modal with count input + optional label prefix; shows progress |
| Bulk redirect update | Confirmation dialog that clearly states what will change |
| Clear individual override | "Revert to batch URL" button with confirmation |
| Download QR code | Direct link to `/api/qrcodes/:id/image?size=1024` as PNG download |
| Delete batch | Confirmation modal warning scan data will be lost |
| Date range on charts | From/to date pickers, default last 30 days |

---

## Tailwind Configuration Notes

- Use `@apply` sparingly — prefer utility classes in templates
- Define a custom colour palette matching your brand in `tailwind.config.ts`
- Use `dark:` variants if dark mode support is desired (optional for v1)

---

## Dev Proxy (Vite)

To avoid CORS issues in development, proxy `/api` and `/r` to the Go backend:

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/r':   'http://localhost:8080'
    }
  }
})
```
