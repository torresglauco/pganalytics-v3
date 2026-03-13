# pgAnalytics Frontend Redesign Specification
## Complete UI/UX Overhaul for Enterprise Observability

**Version:** 1.0.0
**Date:** 2026-03-12
**Status:** Design Phase
**Personas:** DevOps, DBAs, Developers, Executives
**Inspiration:** Datadog, Grafana, New Relic (enterprise observability tools)

---

## 1. Design Philosophy & Vision

### Core Principles
1. **Single Unified View** (Datadog-style) — One integrated dashboard you drill-down into specific areas
2. **Multi-Persona Support** — Same interface serves DevOps, DBAs, Devs, Executives with intelligent defaults
3. **Real-time Observability** — Live updates, instant feedback, responsive interactions
4. **Enterprise Grade** — Professional, trustworthy, scalable visual design
5. **Grafana Integration** — Seamlessly embed Grafana dashboards using Embedded SDK

### Visual Identity
- **Color Palette**: Professional Blue + Slate + Semantic Status Colors
- **Typography**: Clean system fonts with clear hierarchy
- **Density**: Medium density — professional but not cramped
- **Motion**: Subtle, purposeful animations (150-300ms)
- **Dark Mode**: Full support (default light, toggle available)
- **Accessibility**: WCAG AA compliant throughout

---

## 2. Application Architecture

### Top-Level Structure
```
pgAnalytics (Main App)
├── Authentication Layer
│   ├── Login Page
│   ├── Signup Page
│   ├── MFA Verification
│   └── Password Management
│
├── Main Application (Post-Auth)
│   ├── Header (Global Navigation + Search + Notifications)
│   ├── Sidebar (Main Navigation + Shortcuts)
│   └── Main Content Area
│       ├── Dashboard (Home) — Unified drill-down view
│       ├── Logs Viewer — PostgreSQL logs exploration
│       ├── Metrics — Analytics & performance charts
│       ├── Alerts — Alert rules & incident management
│       ├── Grafana — Embedded dashboards
│       ├── Collectors — Manage data sources
│       ├── Channels — Notification configuration
│       ├── Users — Team management (admin)
│       └── Settings — System configuration
│
└── Common Components
    ├── Modals & Dialogs
    ├── Toast Notifications
    ├── Loading States
    └── Error Boundaries
```

---

## 3. Page-Level Designs

### 3.1 Authentication Pages

#### Login Page
**Purpose:** Secure entry point for all users

**Layout:**
```
┌─────────────────────────────────┐
│                                 │
│     pgAnalytics                 │
│     PostgreSQL Observability     │
│                                 │
│     ┌───────────────────────┐   │
│     │ Email                 │   │
│     │ [________________]    │   │
│     │                       │   │
│     │ Password              │   │
│     │ [________________]    │   │
│     │                       │   │
│     │ [Remember Me] ☐       │   │
│     │                       │   │
│     │ [Sign In] (Primary)   │   │
│     │                       │   │
│     │ Don't have an account?│   │
│     │ [Sign Up] (Link)      │   │
│     └───────────────────────┘   │
│                                 │
│     [Login with SSO ▼]          │
│                                 │
└─────────────────────────────────┘
```

**Features:**
- Clean, centered card layout
- Email + Password fields
- "Remember Me" checkbox
- Sign Up link for new users
- SSO option (LDAP/OAuth)
- Error messaging
- Loading state on submit

#### Signup Page
**Purpose:** New user registration

**Layout:**
- Similar card structure to login
- Fields: Email, Name, Password, Confirm Password
- Organization selection (if multi-org)
- Terms acceptance checkbox
- "Already have an account? [Sign In]" link

#### MFA Verification Page
**Purpose:** Multi-factor authentication

**Layout:**
- Card with "Enter Verification Code"
- 6-digit code input (masked, can be numbers/letters)
- "Didn't receive code? [Resend]"
- "Can't access MFA device? [Backup Codes]"

#### Password Reset/Change Page
**Purpose:** Manage password security

**Layout:**
- Current password field
- New password field
- Confirm password field
- Password strength indicator
- Requirements checklist

---

### 3.2 Main Application Layout

#### Header Component
```
┌─────────────────────────────────────────────────────────────┐
│ pgAnalytics  [Search: ________________]  🔔 ⚙ 👤 ▼          │
└─────────────────────────────────────────────────────────────┘
```

**Elements:**
- **Logo/Brand** (left) — clickable to home
- **Search Bar** (center) — search logs, alerts, metrics
- **Notification Bell** (right) — dropdown with recent alerts
- **Settings Icon** (right) — quick access to settings
- **User Menu** (right) — avatar + dropdown (profile, logout)

**Features:**
- Search supports quick commands (e.g., "/metrics", "/logs")
- Real-time notification indicator
- Dark/light mode toggle in menu
- User name + role visible in dropdown

#### Sidebar Navigation
```
┌──────────────────┐
│ Shortcuts        │
│ ▲ Home           │
│ 📊 Dashboard     │
│ 📋 Logs          │
│ 📈 Metrics       │
│ 🚨 Alerts        │
├──────────────────┤
│ Main             │
│ 📁 Collectors    │
│ 🔔 Channels      │
│ ⚙ Grafana        │
├──────────────────┤
│ Admin            │
│ 👥 Users         │
│ ⚙ Settings       │
└──────────────────┘
```

**Features:**
- Collapsible on mobile
- Active page highlighted (blue background + white text)
- Sections (Shortcuts, Main, Admin)
- Icons + labels for clarity
- Hover state (slight highlight)
- Mini-collapse option (show icons only on desktop)

---

### 3.3 Dashboard (Home) — Unified Drill-Down View

**Purpose:** Single entry point for all personas, drill-down to specific areas

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Dashboard                           [Time: Last 24h ▼]     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ SYSTEM HEALTH (Metric Cards Row)                            │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│ │ Collectors   │ │ Active       │ │ Errors       │         │
│ │ OK: 12       │ │ Alerts: 3    │ │ Last 24h: 145│         │
│ │ Down: 0      │ │ Critical: 1  │ │ ↑ 12% trend  │         │
│ └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                              │
│ RECENT ACTIVITY (What Happened?)                            │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 🔴 2m ago  | High error rate detected (prod-db-1)       │ │
│ │ 🟡 15m ago | Slow query detected (>5s)                  │ │
│ │ 🟢 1h ago  | Backup completed successfully              │ │
│ │ [View All Logs ▶]                                       │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ DRILL-DOWN OPTIONS (Click to explore deeper)               │
│ ┌────────────────────────────────────────────────────────┐ │
│ │ 📋 View Logs                                            │ │
│ │ Filter and analyze PostgreSQL logs in detail            │ │
│ │                                                          │ │
│ │ 📈 View Metrics & Analytics                             │ │
│ │ Charts, performance trends, error distribution          │ │
│ │                                                          │ │
│ │ 🚨 View Active Alerts                                   │ │
│ │ See all triggered alerts and manage incidents           │ │
│ │                                                          │ │
│ │ 📊 View Grafana Dashboards                              │ │
│ │ Custom dashboards from Grafana (embedded)               │ │
│ └────────────────────────────────────────────────────────┘ │
│                                                              │
│ COLLECTOR STATUS (Real-time)                                │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Hostname          │ Status  │ Last Seen │ Errors        │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ prod-db-1.aws     │ ✓ OK    │ 2s ago    │ 23 (24h)      │ │
│ │ staging-db.local  │ ⚠ Slow  │ 1m ago    │ 5 (24h)       │ │
│ │ dev-db-local      │ ✗ Down  │ 2h ago    │ — (offline)   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **System Health Cards** — Overview of key metrics (collectors, alerts, errors)
- **Recent Activity Feed** — Last 10 events (logs, alerts, system events)
- **Drill-Down Cards** — Large buttons to navigate to Logs, Metrics, Alerts, Grafana
- **Collector Status Table** — Real-time status of all data sources
- **Time Range Selector** — Global filter affecting all metrics
- **Responsive** — Stack vertically on mobile

---

### 3.4 Logs Viewer Page

**Purpose:** Explore PostgreSQL logs with powerful filtering and search

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Logs                                  [Time: Last 24h ▼]   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ FILTERS & SEARCH                                            │
│ [Instance: All ▼] [Level: All ▼] [User: ______] [Search] │
│ [+ More Filters ▼]                                          │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ LOGS TABLE (Main Content)                                   │
│ ┌────┬─────────┬──────────┬─────────────────────────────┐  │
│ │    │ LEVEL   │ TIME     │ MESSAGE                     │  │
│ ├────┼─────────┼──────────┼─────────────────────────────┤  │
│ │ 📌 │ 🔴 ERROR│ 3 min    │ Connection timeout...       │  │
│ │    │         │ ago      │ [View Details ▶]            │  │
│ ├────┼─────────┼──────────┼─────────────────────────────┤  │
│ │    │ 🟡 WARN │ 5 min    │ Slow query detected...      │  │
│ │    │         │ ago      │ [View Details ▶]            │  │
│ ├────┼─────────┼──────────┼─────────────────────────────┤  │
│ │    │ 🟢 INFO │ 8 min    │ Checkpoint completed        │  │
│ │    │         │ ago      │ [View Details ▶]            │  │
│ └────┴─────────┴──────────┴─────────────────────────────┘  │
│                                                              │
│ Showing 1-50 of 1,234 logs                                  │
│ [< Previous] [1] [2] [3] ... [Next >]                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Filter Bar** — Instance, Level, User, search box + expandable advanced filters
- **Sortable Table** — Click column headers to sort
- **Inline Details** — Click "[View Details ▶]" to expand row
- **Pagination** — 50 rows/page with prev/next/jump to page
- **Pin Button** — Pin important logs to top
- **Real-time Badge** — "● LIVE" indicator with auto-refresh toggle
- **Export** — CSV/JSON export button

#### Log Details (Expanded Row / Modal)
```
┌─────────────────────────────────────────────────────────────┐
│ LOG #4521 DETAILS                                      [✕] │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ METADATA                                                    │
│ Level: 🔴 ERROR  │  Timestamp: 2026-03-12 15:34:22 UTC    │
│ Instance: prod-db-1  │  Database: analytics_prod           │
│ User: app_user  │  Session: sess_abc123                    │
│                                                              │
│ ERROR DETAILS                                               │
│ Code: 42P01  │  Message: relation "events_old" doesn't...  │
│ Hint: Did you mean table "events"?                          │
│ Context: During execution of trigger function              │
│                                                              │
│ QUERY                                                       │
│ SELECT * FROM events_old WHERE id = 123;                   │
│ [Copy] [Show Similar Logs ▶]                               │
│                                                              │
│ RAW LOG                                                     │
│ 2026-03-12 15:34:22 UTC [12345] ERROR: relation...        │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ [Close] [Create Alert] [View Similar]                       │
└─────────────────────────────────────────────────────────────┘
```

---

### 3.5 Metrics & Analytics Page

**Purpose:** Visual analysis of PostgreSQL performance and logs

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Metrics & Analytics                   [Time: Last 24h ▼]   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ KEY METRICS (Row 1)                                         │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│ │ Total Errors │ │ Avg Response │ │ Unique Users │         │
│ │ 1,234        │ │ 234ms        │ │ 45           │         │
│ │ ↑ 12% (24h)  │ │ ↑ 5% (24h)   │ │ ↓ 2% (24h)   │         │
│ └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                              │
│ CHARTS (Row 2 - Side by side)                               │
│ ┌─────────────────────────┐ ┌──────────────────────────┐   │
│ │ ERROR TREND (24h)       │ │ LOG DISTRIBUTION BY LVL  │   │
│ │                         │ │                          │   │
│ │     ▲ Errors            │ │ ERROR    ████ 45%       │   │
│ │     │   ___   ___       │ │ WARNING  ██   20%       │   │
│ │     │  /   \_/   \      │ │ INFO     ██   25%       │   │
│ │ ____│_/           \_    │ │ DEBUG    █    10%       │   │
│ │     └─────────────────→ │ │                          │   │
│ │     Time                │ │                          │   │
│ │ Peak: 15:34 (234 errs)  │ │                          │   │
│ └─────────────────────────┘ └──────────────────────────┘   │
│                                                              │
│ CHARTS (Row 3 - Side by side)                               │
│ ┌─────────────────────────┐ ┌──────────────────────────┐   │
│ │ TOP ERROR CODES         │ │ HOURLY REQUEST VOLUME    │   │
│ │                         │ │                          │   │
│ │ 42P01  ████████ 45%     │ │ ▲ Requests/Hour         │   │
│ │ 08006  ██████   30%     │ │ │     _______           │   │
│ │ 57014  ██       10%     │ │ │    /       \   ___    │   │
│ │ Others █        15%     │ │ │___/         \_/   \   │   │
│ │                         │ │ └──────────────────────→ │   │
│ │                         │ │ Avg: 2,341 req/h        │   │
│ └─────────────────────────┘ └──────────────────────────┘   │
│                                                              │
│ RESPONSE TIME PERCENTILES                                   │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ P50: 145ms │ P75: 234ms │ P95: 567ms │ P99: 1.2s       │ │
│ │ ▓▓▓▓▓▓▓░░░░ 85% under 300ms target                     │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Metric Cards** — Top 3 KPIs with trend indicators
- **Interactive Charts** — Hover for tooltips, click to drill-down
- **Multiple Chart Types** — Line, Bar, Pie charts using Recharts
- **Time Range Selector** — Global 24h, 7d, 30d, custom
- **Chart Customization** — Toggle series, change aggregation
- **Export** — Download chart as PNG or data as CSV

---

### 3.6 Alerts Page

**Purpose:** Manage alert rules and view active incidents

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Alerts                                [+ New Alert]         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ ACTIVE INCIDENTS (Top)                                      │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 🔴 CRITICAL: High Error Rate (prod-db-1)               │ │
│ │ Started 5 min ago • 245 errors in last 5 min           │ │
│ │ Alert: "Error rate > 100/5min"                         │ │
│ │ Actions: [Acknowledge] [Snooze 1h] [View Logs]         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ ALERT RULES (Configured Rules)                              │
│ ┌────┬──────────────────────┬────────┬──────┬────────────┐  │
│ │    │ NAME                 │ STATUS │ HIT  │ LAST SENT  │  │
│ ├────┼──────────────────────┼────────┼──────┼────────────┤  │
│ │ ☐  │ High Error Rate      │ ✓ On   │ 3x   │ 5 min ago  │  │
│ │    │ Trigger: >100 errs/5 │        │      │            │  │
│ │    │ [Edit] [Test] [...]  │        │      │            │  │
│ ├────┼──────────────────────┼────────┼──────┼────────────┤  │
│ │ ☐  │ Slow Query Alert     │ ✓ On   │ 1x   │ 30 min ago │  │
│ │    │ Trigger: query > 5s  │        │      │            │  │
│ │    │ [Edit] [Test] [...]  │        │      │            │  │
│ ├────┼──────────────────────┼────────┼──────┼────────────┤  │
│ │ ☐  │ Collector Down       │ ○ Off  │ —    │ —          │  │
│ │    │ Trigger: no heartbeat│        │      │            │  │
│ │    │ [Edit] [Test] [...]  │        │      │            │  │
│ └────┴──────────────────────┴────────┴──────┴────────────┘  │
│                                                              │
│ Showing 3 of 12 rules                                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Active Incidents Section** — Highest priority at top
- **Alert Rules Table** — All configured rules with status toggle
- **Quick Actions** — Edit, Test, Delete, View History
- **Filters** — By status (On/Off), by type, by channel
- **Bulk Actions** — Select multiple rules to enable/disable/delete

#### Create/Edit Alert Modal
```
┌─────────────────────────────────────────────────────────┐
│ Create Alert Rule                                  [✕]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Name *                                                  │
│ [High Error Rate                              ]         │
│                                                          │
│ Description                                             │
│ [When error logs exceed threshold in time window]       │
│                                                          │
│ Condition *                                             │
│ When: [ERROR logs ▼] [> 100 ▼] in [5 minutes ▼]       │
│ [+ Add Condition]                                       │
│                                                          │
│ Notification Channels *                                 │
│ ☑ Email (dev-team@company.com)                         │
│ ☑ Slack (#alerts)                                       │
│ ☐ PagerDuty                                             │
│ [+ Add Channel]                                         │
│                                                          │
│ Frequency Limits                                        │
│ ☑ Max [1] alerts per [1 hour]                          │
│                                                          │
│ Test Alert                                              │
│ [Send Test Alert]                                       │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ [Cancel] [Save Alert]                                   │
└─────────────────────────────────────────────────────────┘
```

---

### 3.7 Collectors Management Page

**Purpose:** Register and manage data source collectors

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Collectors                            [+ Register Collector]│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ COLLECTOR SUMMARY (Metric Cards)                            │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│ │ Total        │ │ Healthy      │ │ Down         │         │
│ │ 12           │ │ 11           │ │ 1            │         │
│ └──────────────┘ └──────────────┘ └──────────────┘         │
│                                                              │
│ COLLECTORS TABLE                                            │
│ ┌────┬──────────────────┬────────┬──────────┬────────────┐  │
│ │ ID │ HOSTNAME         │ STATUS │ LAST     │ LOGS       │  │
│ ├────┼──────────────────┼────────┼──────────┼────────────┤  │
│ │ 1  │ prod-db-1.aws    │ ✓ OK   │ 2s ago   │ 245 (24h)  │  │
│ │    │ Environment: prod │        │          │            │  │
│ │    │ [View Details] [Edit] [Logs] [Delete]             │  │
│ ├────┼──────────────────┼────────┼──────────┼────────────┤  │
│ │ 2  │ staging-db.local │ ⚠ Slow │ 1m ago   │ 12 (24h)   │  │
│ │    │ Environment: staging                              │  │
│ │    │ [View Details] [Edit] [Logs] [Delete]             │  │
│ ├────┼──────────────────┼────────┼──────────┼────────────┤  │
│ │ 3  │ dev-db-local     │ ✗ Down │ 2h ago   │ — (offline)│  │
│ │    │ Environment: dev                                   │  │
│ │    │ [View Details] [Edit] [Logs] [Delete]             │  │
│ └────┴──────────────────┴────────┴──────────┴────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Collector Summary** — Total, healthy, down counts
- **Status Table** — Hostname, environment, last heartbeat, error count
- **Quick Actions** — View details, edit, view logs, delete
- **Bulk Actions** — Select multiple to disable/delete

#### Register Collector Modal
```
┌─────────────────────────────────────────────────────────┐
│ Register Collector                                 [✕]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Hostname *                                              │
│ [prod-db-1.region.rds.amazonaws.com] [Test Connection]│
│                                                          │
│ Environment *                                           │
│ [Production ▼]                                          │
│                                                          │
│ Group                                                   │
│ [AWS-RDS ▼]                                             │
│                                                          │
│ Description                                             │
│ [Optional description...]                               │
│                                                          │
│ Registration Secret *                                   │
│ [••••••••••••••] [Show]                                │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ [Cancel] [Register]                                     │
│                                                          │
│ Success Response (if registered):                       │
│ ✓ Collector Registered!                                │
│ ID: col_abc123def456                                    │
│ Token: [Copy]                                           │
│ [Register Another]                                      │
└─────────────────────────────────────────────────────────┘
```

---

### 3.8 Notification Channels Page

**Purpose:** Configure where alerts get sent (Email, Slack, PagerDuty, etc.)

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Notification Channels                [+ Add Channel]        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ CHANNELS TABLE                                              │
│ ┌────┬──────────┬──────────────────┬────────┬────────────┐  │
│ │ ID │ TYPE     │ DESTINATION      │ STATUS │ ACTIONS    │  │
│ ├────┼──────────┼──────────────────┼────────┼────────────┤  │
│ │ 1  │ 📧 Email │ dev@company.com  │ ✓ Ok   │ [E][T][D]  │  │
│ ├────┼──────────┼──────────────────┼────────┼────────────┤  │
│ │ 2  │ 💬 Slack │ #alerts          │ ✓ Ok   │ [E][T][D]  │  │
│ ├────┼──────────┼──────────────────┼────────┼────────────┤  │
│ │ 3  │ 📟 PagerDuty│ Service: PgDB  │ ✓ Ok   │ [E][T][D]  │  │
│ ├────┼──────────┼──────────────────┼────────┼────────────┤  │
│ │ 4  │ 🪝 Webhook│ https://...      │ ✗ Fail │ [E][T][D]  │  │
│ └────┴──────────┴──────────────────┴────────┴────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Channel Type Icons** — Email, Slack, PagerDuty, Webhook, Jira
- **Status Indicator** — Ok/Failed with last test time
- **Quick Actions** — Edit, Test, Delete

#### Channel Configuration Modal
```
Different forms based on channel type:

EMAIL:
- Email address
- Subject template
- Message template

SLACK:
- Webhook URL
- Channel
- Message format

PAGERDUTY:
- API Key
- Service ID
- Severity mapping

WEBHOOK:
- URL
- HTTP method (POST/PUT)
- Custom headers
- Payload template

JIRA:
- Instance URL
- API key
- Project key
- Issue type
```

---

### 3.9 Grafana Dashboards Page

**Purpose:** Embedded Grafana dashboards using Embedded SDK

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Grafana Dashboards                   [+ Add Dashboard]      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ AVAILABLE DASHBOARDS (Grid of Cards)                        │
│                                                              │
│ ┌──────────────────────┐ ┌──────────────────────┐           │
│ │ 📊 PostgreSQL Metrics│ │ 📊 Query Performance  │           │
│ │ Updated: 5 min ago   │ │ Updated: 2 min ago   │           │
│ │ [View] [Settings]    │ │ [View] [Settings]    │           │
│ └──────────────────────┘ └──────────────────────┘           │
│                                                              │
│ ┌──────────────────────┐ ┌──────────────────────┐           │
│ │ 📊 Connections       │ │ 📊 Replication Status │           │
│ │ Updated: 1 min ago   │ │ Updated: 30s ago     │           │
│ │ [View] [Settings]    │ │ [View] [Settings]    │           │
│ └──────────────────────┘ └──────────────────────┘           │
│                                                              │
└─────────────────────────────────────────────────────────────┘

WHEN CLICKING "VIEW":

┌─────────────────────────────────────────────────────────────┐
│ PostgreSQL Metrics                          [← Back] [Settings]│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ [Grafana Dashboard Embedded Using SDK]                      │
│                                                              │
│ - Full Grafana dashboard rendered inline                    │
│ - User can interact (drill-down, change time range)         │
│ - Respects Grafana permissions                              │
│ - Auto-refresh based on dashboard settings                  │
│                                                              │
│ [Grafana content here - height: 100vh, width: 100%]         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Dashboard Grid** — Card view of available dashboards
- **Embedded View** — Full-screen Grafana dashboard using Embedded SDK
- **Time Sync** — Dashboard respects global time range selector
- **Quick Access** — Cards show last update time
- **Settings** — Configure which Grafana dashboards to show

---

### 3.10 User Management Page (Admin)

**Purpose:** Manage team members and permissions

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Users                                 [+ Invite User]       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ USERS TABLE                                                  │
│ ┌────┬──────────────────┬──────────┬────────┬────────────┐  │
│ │    │ NAME             │ EMAIL    │ ROLE   │ STATUS     │  │
│ ├────┼──────────────────┼──────────┼────────┼────────────┤  │
│ │ ☐  │ John Doe         │ john@... │ Admin  │ ✓ Active   │  │
│ │    │ [Edit] [Disable] │          │        │            │  │
│ ├────┼──────────────────┼──────────┼────────┼────────────┤  │
│ │ ☐  │ Jane Smith       │ jane@... │ Editor │ ✓ Active   │  │
│ │    │ [Edit] [Disable] │          │        │            │  │
│ ├────┼──────────────────┼──────────┼────────┼────────────┤  │
│ │ ☐  │ Bob Wilson       │ bob@...  │ Viewer │ ✗ Inactive │  │
│ │    │ [Edit] [Resend]  │          │        │            │  │
│ └────┴──────────────────┴──────────┴────────┴────────────┘  │
│                                                              │
│ Showing 3 of 12 users                                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Users Table** — Name, email, role, status
- **Role Selection** — Admin, Editor, Viewer
- **Status Toggle** — Enable/disable users
- **Bulk Actions** — Invite multiple, change roles in bulk

#### Invite User Modal
```
┌─────────────────────────────────────────────────────────┐
│ Invite User                                        [✕]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Email *                                                 │
│ [user@company.com                              ]        │
│                                                          │
│ Role *                                                  │
│ [Admin ▼]  (Admin, Editor, Viewer)                     │
│                                                          │
│ Message                                                 │
│ [Optional message to include in invitation...]          │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ [Cancel] [Send Invitation]                              │
└─────────────────────────────────────────────────────────┘
```

---

### 3.11 Settings Page

**Purpose:** System configuration and preferences

**Layout:**
```
┌─────────────────────────────────────────────────────────────┐
│ Settings                                                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ TABS: [General] [Organization] [Integrations] [Advanced]   │
│                                                              │
│ TAB: GENERAL                                                │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Theme                                                   │ │
│ │ ○ Light  ○ Dark  ○ Auto                                │ │
│ │                                                          │ │
│ │ Time Zone                                               │ │
│ │ [America/New_York ▼]                                    │ │
│ │                                                          │ │
│ │ Date Format                                             │ │
│ │ [MM/DD/YYYY ▼]                                          │ │
│ │                                                          │ │
│ │ Notifications                                           │ │
│ │ ☑ Email notifications                                   │ │
│ │ ☑ Browser notifications                                │ │
│ │ ☑ Sound alerts                                          │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ TAB: ORGANIZATION                                           │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Organization Name                                       │ │
│ │ [My Company Inc.                              ]         │ │
│ │                                                          │ │
│ │ Logo                                                    │ │
│ │ [📎 Upload Logo]                                        │ │
│ │                                                          │ │
│ │ Billing Email                                           │ │
│ │ [billing@company.com                        ]          │ │
│ │                                                          │ │
│ │ Danger Zone                                             │ │
│ │ [Delete Organization]                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ TAB: INTEGRATIONS                                           │
│ [Grafana] [DataDog] [Custom Webhooks] [API Keys]           │
│                                                              │
│ TAB: ADVANCED                                               │
│ [Logs Retention] [API Configuration] [SSO Settings]        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Tabbed Interface** — Organized into logical sections
- **Theme Toggle** — Light/Dark/Auto support
- **Notification Preferences** — Control how users are notified
- **Organization Settings** — Company name, logo, billing
- **Integration Settings** — Grafana API keys, custom webhooks
- **Danger Zone** — Destructive actions clearly marked

---

## 4. Component Library

### Base Components (Reusable)
- **Button** — Primary, Secondary, Danger variants
- **Input** — Text, email, password, search fields
- **Select** — Single and multi-select dropdowns
- **Textarea** — Multi-line text input
- **Checkbox & Radio** — Form controls
- **Card** — Container for grouped content
- **Badge** — Status and label indicators
- **Modal** — Dialog boxes for actions
- **Tooltip** — Contextual help text
- **Loading Spinner** — Async operation feedback
- **Toast** — Dismissible notifications
- **Pagination** — Page navigation
- **Tabs** — Section navigation
- **Dropdown Menu** — Context menus

### Composite Components
- **MetricCard** — Display KPI with trend
- **StatusBadge** — Status indicator (ok, warning, error)
- **DataTable** — Sortable, paginated table
- **LineChart** — Time-series visualization
- **BarChart** — Comparison visualization
- **PieChart** — Distribution visualization
- **FilterBar** — Grouped filters
- **DetailPanel** — Expandable detail view
- **ConfirmDialog** — Confirm destructive actions

---

## 5. Color Palette & Theming

### Light Mode (Default)
```css
/* Primary */
--primary-50:   #EFF6FF
--primary-100:  #DBEAFE
--primary-500:  #3B82F6
--primary-600:  #2563EB
--primary-700:  #1D4ED8

/* Semantic */
--success:      #10B981
--warning:      #F59E0B
--error:        #EF4444
--info:         #3B82F6

/* Neutral */
--gray-50:      #F9FAFB
--gray-100:     #F3F4F6
--gray-200:     #E5E7EB
--gray-500:     #6B7280
--gray-900:     #111827

/* Backgrounds */
--bg-primary:   #FFFFFF
--bg-secondary: #F9FAFB
```

### Dark Mode
- Invert brightness while maintaining saturation
- Use Slate-900 for primary background
- Ensure 4.5:1 contrast on text

### Theme Implementation
- Tailwind CSS configuration
- CSS variables for dynamic theming
- Support for system color scheme preference

---

## 6. Navigation & Routing

### Main Routes
```
/                    Dashboard (Home)
/logs                Logs Viewer
/metrics             Metrics & Analytics
/alerts              Alert Rules & Incidents
/grafana             Grafana Dashboards
/collectors          Collector Management
/channels            Notification Channels
/users               User Management (Admin)
/settings            Settings & Configuration

/login               Authentication
/signup              User Registration
/verify-mfa          MFA Verification
/reset-password      Password Reset
```

### Sidebar Navigation
- Shortcuts section (Home, Logs, Metrics, Alerts)
- Main section (Collectors, Channels, Grafana)
- Admin section (Users, Settings)
- Collapsible on mobile

---

## 7. State Management & Data Flow

### Global State (Zustand)
- User authentication
- Current user profile
- Organization context
- Theme preference
- Notification center
- Time range selection (for charts/logs)

### Local State (React Hooks)
- Form inputs
- UI toggles (filters, modals, etc.)
- Loading/error states
- Pagination state

### Data Fetching (React Query / TanStack)
- Logs queries (paginated)
- Metrics aggregations
- Alert rules CRUD
- User management
- Settings persistence

---

## 8. Responsive Design

### Breakpoints
- **Mobile** (0-375px) — Single column, collapsed sidebar, touch-friendly
- **Tablet** (376-768px) — Two columns, sidebar toggleable
- **Desktop** (769-1440px) — Full layout, visible sidebar
- **Large** (1441px+) — Max-width container, comfortable spacing

### Mobile Optimizations
- Vertical stack layouts
- Full-width modals
- Bottom sheet for filters
- Touch targets ≥44px
- Swipe navigation support

---

## 9. Accessibility & Compliance

### WCAG AA Compliance
- Color contrast 4.5:1 for text
- Keyboard navigation (Tab, Enter, Escape)
- Screen reader support (ARIA labels)
- Focus indicators (2px blue outline)
- Semantic HTML (form, nav, main, etc.)

### Keyboard Shortcuts
- `/` — Open search
- `?` — Show keyboard shortcuts help
- `Escape` — Close modals/dropdowns
- `Ctrl+K` — Command palette (future)

---

## 10. Performance Targets

### Core Web Vitals
- **LCP**: < 2.5s
- **FID**: < 100ms
- **CLS**: < 0.1

### Optimization Strategies
- Code splitting by route
- Image optimization (WebP/AVIF)
- Lazy loading below-fold content
- Virtual scrolling for large tables
- API response caching (5-10min TTL)

---

## 11. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [x] Design System finalized
- [ ] Project setup (React 18, TypeScript, Tailwind)
- [ ] Base components library
- [ ] Authentication flows
- [ ] Main layout (Header, Sidebar)
- [ ] Dashboard (home page)

### Phase 2: Core Features (Week 3-4)
- [ ] Logs Viewer page
- [ ] Metrics & Analytics page
- [ ] Alerts management
- [ ] Collectors management

### Phase 3: Integrations & Admin (Week 5-6)
- [ ] Grafana embedded dashboards
- [ ] Notification channels
- [ ] User management
- [ ] Settings page

### Phase 4: Polish & Testing (Week 7-8)
- [ ] Dark mode support
- [ ] Mobile responsiveness
- [ ] Accessibility audit
- [ ] Performance optimization
- [ ] End-to-end testing
- [ ] Documentation & Storybook

---

## 12. Success Criteria

✅ Single unified dashboard (drill-down pattern)
✅ Multi-persona support (DevOps, DBA, Dev, Executive)
✅ Professional/corporate aesthetic (Datadog-inspired)
✅ Real-time data updates
✅ Grafana integration via Embedded SDK
✅ Full WCAG AA accessibility
✅ Mobile responsive (375px+)
✅ Dark mode support
✅ < 2.5s LCP, < 100ms FID
✅ All 10 pages redesigned and functional

---

**Design Status:** ✅ Complete & Ready for Implementation
**Next Step:** Present to user for approval, then begin Phase 1
