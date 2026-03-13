# pgAnalytics Frontend Design System
## Professional/Corporate Analytics Dashboard

**Version:** 3.3.0
**Last Updated:** 2026-03-12
**Stack:** React 18 + TypeScript + Tailwind CSS + Recharts

---

## 1. Design Principles

### Primary Style: Professional/Corporate
- **Visual Hierarchy**: Bold headers, organized sections, clear information structure
- **Color Scheme**: Professional blue/slate palette with semantic reds (errors), yellows (warnings), greens (success)
- **Density**: Medium density — comfortable spacing without excessive whitespace
- **Typography**: Clear hierarchy with system fonts (Inter/SF Pro for professional feel)
- **Motion**: Subtle, purposeful animations (150-300ms) to convey state changes
- **Accessibility**: WCAG AA compliant (4.5:1 contrast minimum), keyboard navigation, screen reader support

### Key Aesthetic Features
- Clean card-based layouts with consistent elevation (shadows)
- Data-driven design (charts, metrics, tables prioritized)
- Real-time indicators and live update badges
- Professional color semantics (red=error, yellow=warning, green=success, blue=info)
- Responsive from 375px (mobile) to 1920px (desktop)

---

## 2. Color Palette

### Semantic Colors
```
Primary:      #0066CC (Blue-600)     — Main actions, links
Secondary:    #0052A3 (Blue-700)     — Secondary buttons, hover states
Success:      #10B981 (Emerald-500)  — Success messages, OK status
Warning:      #F59E0B (Amber-500)    — Warning logs, caution states
Error:        #EF4444 (Red-500)      — Errors, critical logs, danger actions
Info:         #3B82F6 (Blue-500)     — Informational messages
Neutral:      #6B7280 (Gray-500)     — Secondary text, disabled states

Background:   #FFFFFF (White)        — Light theme
Surface:      #F9FAFB (Gray-50)      — Card/container backgrounds
Border:       #E5E7EB (Gray-200)     — Dividers, input borders
Text Primary: #111827 (Gray-900)     — Headings, body text
Text Secondary: #6B7280 (Gray-500)   — Labels, helper text
```

### Status Badges
- **Active/OK**: bg-emerald-100, text-emerald-800, border-emerald-300
- **Warning**: bg-amber-100, text-amber-800, border-amber-300
- **Error**: bg-red-100, text-red-800, border-red-300
- **Neutral**: bg-gray-100, text-gray-800, border-gray-300

### Dark Mode (Future)
All colors should support dark mode variants using Tailwind dark: prefix.
Strategy: Desaturated/lighter tonal variants, not inverted.

---

## 3. Typography System

### Font Stack
```typescript
Font Family: "Inter", "system-ui", "-apple-system", sans-serif
Fallback: System fonts for iOS/Android compatibility
```

### Type Scale & Hierarchy
```
Display:    32px / 600 weight   — Page titles (Dashboard, Analytics)
Heading 1:  24px / 600 weight   — Card titles, section headers
Heading 2:  20px / 600 weight   — Subsection headers
Heading 3:  18px / 500 weight   — Form labels, table headers
Body:       16px / 400 weight   — Primary text content
Small:      14px / 400 weight   — Secondary text, descriptions
Label:      12px / 500 weight   — Input labels, captions, badges
Mono:       13px / 400 weight   — Code, timestamps, numbers
```

### Line Height & Spacing
- Body text: 1.5 (24px)
- Headings: 1.2 (28-38px)
- Form inputs: 1.5
- Code: 1.6

### Number Formatting
- Use **tabular/monospaced** figures for data columns (timestamps, counts, prices)
- Locale-aware formatting (commas, decimals based on region)

---

## 4. Spacing & Layout System

### Spacing Scale (4dp/8dp increments)
```
xs:   4px     (tight spacing)
sm:   8px     (component padding)
md:   12px    (section padding)
lg:   16px    (card spacing)
xl:   24px    (vertical section spacing)
2xl:  32px    (large gaps)
3xl:  48px    (hero section spacing)
```

### Layout Guidelines
- **Max-width**: 1440px (desktop), 100% (mobile)
- **Horizontal insets**: 16px (mobile), 24px (tablet), 32px (desktop)
- **Grid system**: 12-column responsive grid
- **Card spacing**: 16px gap between cards
- **Section spacing**: 24-32px vertical gaps

### Responsive Breakpoints
```
Mobile:      0-375px   (full-width, single column)
Tablet:      376-768px (2 columns, adjusted spacing)
Desktop:     769-1440px (3+ columns, max-width container)
Large:       1441px+   (full-width grid layout)
```

---

## 5. Component Specifications

### 5.1 Buttons

#### Primary Button
- **State**: Default bg-blue-600, hover:bg-blue-700, disabled:opacity-50
- **Size**: Height 44px (minimum touch target), padding-x 16px
- **Text**: 14px font-semibold, white color
- **Border**: None (solid fill)
- **Feedback**: 150ms transition, cursor-pointer

#### Secondary Button
- **State**: bg-gray-200, hover:bg-gray-300, disabled:opacity-50
- **Size**: 44px height, 16px padding-x
- **Border**: 1px gray-300
- **Icon support**: Left-aligned icon with 8px spacing

#### Danger Button
- **State**: bg-red-600, hover:bg-red-700, text-white
- **Usage**: Delete, destructive actions
- **Confirmation**: Must include confirmation modal

#### Button Accessibility
- Minimum size: 44×44px (touch)
- Focus ring: 2px blue-500, 2px offset
- Aria-label for icon-only buttons
- Disabled attribute (not just opacity)

### 5.2 Forms

#### Input Field
- **Height**: 40px
- **Padding**: 8px 12px
- **Border**: 1px gray-300, focus:border-blue-500 + 2px focus ring
- **Label**: Above input, required fields marked with red asterisk (*)
- **Helper text**: Below input in 12px gray-600
- **Placeholder**: Visible but not replacement for label
- **Error state**: Red border, error message in red-600 below field

#### Select/Dropdown
- **Styling**: Same as input (40px height, gray-300 border)
- **Chevron icon**: Right-aligned, gray-400
- **Options**: Keyboard accessible (arrow keys), searchable for 5+ items
- **Mobile**: System picker on iOS/Android

#### Textarea
- **Min height**: 100px
- **Resizable**: Vertical only
- **Styled**: Same border/focus behavior as input

#### Checkbox & Radio
- **Size**: 18×18px (touch target)
- **Spacing**: 8px gap between label and control
- **Focus**: 2px focus ring around control
- **Label**: Clickable (wraps <label> element)

### 5.3 Tables

#### Table Structure
- **Header**: bg-gray-50, font-semibold 14px, border-bottom 1px gray-200
- **Row height**: 44px (minimum)
- **Cell padding**: 12px horizontal, 10px vertical
- **Zebra striping**: Optional (bg-gray-50 on even rows for better readability)
- **Hover state**: bg-gray-100 (subtle highlight)
- **Border**: 1px gray-200 between rows and edges

#### Column Alignment
- **Text**: Left-aligned (default)
- **Numbers**: Right-aligned (tabular figures)
- **Timestamps**: Center or right-aligned
- **Actions**: Right-aligned

#### Sorting & Filtering
- **Sort indicator**: Up/down chevron next to header text
- **Active sort**: Bold header, blue chevron
- **Pagination**: Below table, shows "Showing X–Y of Z" + page buttons
- **Responsive**: Horizontal scroll on mobile (not full table reflow)

### 5.4 Cards

#### Card Container
- **Background**: White (#FFFFFF)
- **Border**: 1px gray-200
- **Border-radius**: 6px
- **Padding**: 16px
- **Shadow**: `0 1px 3px rgba(0,0,0,0.1)` (subtle)
- **Hover state**: shadow-md (slightly elevated on interactive cards)

#### Card Header (optional)
- **Border-bottom**: 1px gray-200
- **Padding**: 16px, margin-bottom: 16px
- **Title**: 18px font-semibold

### 5.5 Badges/Chips

#### Status Badge
- **Padding**: 4px 8px
- **Border-radius**: 4px
- **Font-size**: 12px font-semibold
- **Colors**: Semantic (red/yellow/green/gray)
- **Example**: `<span class="px-2 py-1 text-xs font-semibold rounded bg-red-100 text-red-800">`

#### Priority/Level Badge
- **ERROR**: Red background, bold
- **WARNING**: Yellow background
- **INFO**: Blue background
- **DEBUG**: Gray background (low emphasis)

### 5.6 Modals/Dialogs

#### Modal Container
- **Backdrop**: Scrim with 40-60% opacity (rgba(0,0,0,0.5))
- **Card**: bg-white, rounded-lg, min-width 400px, max-width 90vw
- **Padding**: 24px
- **Close button**: X icon, top-right corner
- **Focus management**: Focus trap (first interactive element on open)
- **Animation**: Fade-in + scale-up (150-200ms)

#### Modal Header
- **Title**: 20px font-semibold
- **Border-bottom**: 1px gray-200 (optional)
- **Close icon**: Top-right, cursor-pointer

#### Modal Body
- **Padding**: 16px vertical, inherits horizontal from container
- **Scrollable**: If content exceeds viewport height
- **Max-height**: 70vh

#### Modal Footer (if applicable)
- **Border-top**: 1px gray-200
- **Padding**: 16px
- **Buttons**: Primary action on right, secondary/cancel on left

### 5.7 Tooltips & Popovers

#### Tooltip
- **Background**: Gray-900 (slate-900)
- **Text**: White, 12px
- **Padding**: 6px 8px
- **Border-radius**: 4px
- **Arrow**: Directional indicator
- **Z-index**: 1000
- **Trigger**: Hover (desktop), tap (mobile)
- **Delay**: 200ms before showing

#### Popover (Menu)
- **Background**: White
- **Border**: 1px gray-200
- **Shadow**: md (4px shadow)
- **Border-radius**: 6px
- **Padding**: 8px
- **Items**: 36px height, 12px padding-x

---

## 6. Page Layouts

### 6.1 Logs Dashboard (Main Page)

```
┌─────────────────────────────────────────────────────────────┐
│ PostgreSQL LOGS VIEWER                        [?] [⚙]      │
│ Production Instance — db.prod.aws.com                      │
├─────────────────────────────────────────────────────────────┤
│ 📊 METRICS OVERVIEW (Horizontal Cards Row)                  │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │
│ │ ERROR LOGS   │ │ WARNING LOGS  │ │ AVG RESPONSE │         │
│ │ 1,234        │ │ 456          │ │ 234ms        │         │
│ │ ↑ 12% (24h)  │ │ ↓ 3% (24h)    │ │ ↑ 5% (24h)   │         │
│ └──────────────┘ └──────────────┘ └──────────────┘         │
├─────────────────────────────────────────────────────────────┤
│ 🔍 FILTERS & CONTROLS                                       │
│ [Level: All ▼] [Time: Last 24h ▼] [Search ____] [Refresh]  │
├─────────────────────────────────────────────────────────────┤
│ 📋 LOGS TABLE (Primary Content)                             │
│ ┌────┬──────────┬───────────┬────────────────────────────┐ │
│ │ ID │ SEVERITY │ TIME      │ MESSAGE                    │ │
│ ├────┼──────────┼───────────┼────────────────────────────┤ │
│ │ #1 │ 🔴 ERROR │ 3 min ago │ Connection timeout to...   │ │
│ │    │          │           │ [View Details →]           │ │
│ ├────┼──────────┼───────────┼────────────────────────────┤ │
│ │ #2 │ 🟡 WARN  │ 5 min ago │ Slow query detected...     │ │
│ │    │          │           │ [View Details →]           │ │
│ └────┴──────────┴───────────┴────────────────────────────┘ │
│                                                              │
│ Showing 1-20 of 1,234 logs                                  │
│ [< Previous] [1] [2] [3] [Next >]                           │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- Header with instance selector
- Metric cards showing key stats (error count, warning count, response time)
- Filter bar with log level, time range, search
- Main table with sortable columns
- Row actions (View Details modal)
- Pagination controls

### 6.2 Log Details Modal

```
┌─────────────────────────────────────────────────────────────┐
│ LOG #1234 DETAILS                                [✕]       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ METADATA                                                    │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Level:       🔴 ERROR                                  │ │
│ │ Timestamp:   2026-03-12 15:34:22 UTC                  │ │
│ │ Instance:    prod-db-1 (192.168.1.100)                │ │
│ │ Database:    analytics_prod                            │ │
│ │ User:        app_user                                  │ │
│ │ Session ID:  sess_abc123def456                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ ERROR DETAILS                                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Error Code:  42P01                                      │ │
│ │ Message:     relation "events_old" does not exist       │ │
│ │ Hint:        Did you mean to reference table "events"?  │ │
│ │ Context:     During execution of trigger function      │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ QUERY DETAILS                                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Query:  SELECT * FROM events_old WHERE id = 123;       │ │
│ │ Hash:   0xabc123def456...                              │ │
│ │ [Copy Query] [Show Similar Logs ▼]                     │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ RAW LOG MESSAGE                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 2026-03-12 15:34:22 UTC [12345] ERROR:  relation       │ │
│ │ "events_old" does not exist during execution of        │ │
│ │ trigger function trigger_archive_old_events            │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ [Close] [Create Alert for This] [View Similar (3)]          │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- Metadata section (level, timestamp, instance, database, user)
- Error details (code, message, hint, context)
- Query details (with copy button)
- Raw log message (scrollable)
- Related actions (create alert, view similar logs)

### 6.3 Filters & Search Panel

```
FILTER PANEL (Sidebar or Expandable)
┌─────────────────────────┐
│ FILTERS                 │
├─────────────────────────┤
│ Log Level               │
│ ☐ DEBUG                 │
│ ☑ INFO                  │
│ ☑ WARNING               │
│ ☑ ERROR                 │
│ ☑ FATAL                 │
├─────────────────────────┤
│ Time Range              │
│ ⊙ Last 24h              │
│ ○ Last 7 days           │
│ ○ Last 30 days          │
│ ○ Custom                │
│   [From:] _________     │
│   [To:]   _________     │
├─────────────────────────┤
│ Error Code              │
│ [Type error code...]    │
├─────────────────────────┤
│ User/Session            │
│ [Type username...]      │
├─────────────────────────┤
│ Database                │
│ [Select Database ▼]     │
├─────────────────────────┤
│ [Apply Filters]         │
│ [Reset]                 │
└─────────────────────────┘
```

**Key Features:**
- Checkbox filters (log level)
- Radio buttons (time range)
- Text search (error code, user)
- Dropdown (database selection)
- Apply/Reset buttons
- Mobile: Collapsible panel or bottom sheet

### 6.4 Analytics Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│ ANALYTICS                              [Time Range: 24h ▼]  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ ERROR TREND (24h)           │  LOG DISTRIBUTION BY LEVEL   │
│ ┌────────────────────────┐  │  ┌──────────────────────────┐│
│ │ ▲ Error Count          │  │  │ ERROR     ████ 45%      ││
│ │ │     ___   ___        │  │  │ WARNING   ██   20%      ││
│ │ │    /   \_/   \       │  │  │ INFO      ██   25%      ││
│ │ │___/           \_     │  │  │ DEBUG     █    10%      ││
│ │ └─────────────────────┘  │  │ └──────────────────────────┘│
│ │ Peak: 15:34 (234 errs)   │  │                            │
│                                                              │
│ TOP ERROR CODES              │  HOURLY ACTIVITY            │
│ ┌──────────────────────────┐ │ ┌──────────────────────────┐│
│ │ 42P01 (table missing)    │ │ │ ▲ Requests/Hour         ││
│ │  ████████████ 45% (234)  │ │ │ │           ___         ││
│ │                          │ │ │ │  _____   /   \   __  ││
│ │ 08006 (connection lost)  │ │ │ │ /     \_/     \_/  \_││
│ │  ████████ 30% (155)      │ │ │ │                      ││
│ │                          │ │ │ └──────────────────────────┘│
│ │ 57014 (statement timeout)│ │ │ Avg: 2,341 req/h          │
│ │  ██ 10% (52)             │ │ │ Peak: 3,245 req/h         │
│ │                          │ │ │                            │
│ │ [Show More ▼]            │ │ └──────────────────────────┘│
│ └──────────────────────────┘                               │
│                                                              │
│ RESPONSE TIME PERCENTILES                                   │
│ ┌─────────────────────────────────────────────────────────┐│
│ │ P50:  145ms   P75: 234ms   P95: 567ms   P99: 1,234ms   ││
│ │ ▓▓▓▓▓▓▓░░░░ (85% under 300ms target)                   ││
│ └─────────────────────────────────────────────────────────┘│
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- Time range selector (24h, 7d, 30d, custom)
- Line chart (error trends)
- Pie chart (log distribution)
- Bar chart (top error codes)
- Line chart (hourly activity)
- Metric cards (response time percentiles)
- All charts are interactive (hover tooltips, click for details)

### 6.5 Email Alerts Configuration

```
┌─────────────────────────────────────────────────────────────┐
│ EMAIL ALERTS CONFIGURATION                    [+ New Alert] │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ EXISTING ALERTS                                             │
│                                                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ALERT: High Error Rate                    [Edit] [Delete]│ │
│ │                                                          │ │
│ │ Trigger:  When ERROR logs > 100 in last 30 minutes     │ │
│ │ Recipients: dev-team@company.com, alerts@company.com   │ │
│ │ Status: ✓ Enabled                                       │ │
│ │ Last sent: 2026-03-12 14:22 UTC                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ALERT: Query Timeout Detected                [Edit] [Delete]│
│ │                                                          │ │
│ │ Trigger:  When log level = FATAL in last 5 minutes     │ │
│ │ Recipients: dba-team@company.com                        │ │
│ │ Status: ○ Disabled                                      │ │
│ │ Last sent: 2026-03-11 08:45 UTC                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ [New Alert Form]                                            │
│                                                              │
│ Alert Name *                                                │
│ [Critical Database Events]                                  │
│                                                              │
│ Condition *                                                 │
│ When: [ERROR logs ▼] [> 50 ▼] in [last 15 minutes ▼]      │
│       [+ Add Condition]                                     │
│                                                              │
│ Recipients *                                                │
│ [dev-team@company.com, ops@company.com]                    │
│ [+ Add Email]                                               │
│                                                              │
│ Email Template                                              │
│ Subject: [PostgreSQL Alert: {{alertName}}]                 │
│ Message: [Default template ▼]                              │
│                                                              │
│ Frequency Limits                                            │
│ ☑ Don't send more than [1] alerts per [1 hour]            │
│                                                              │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ [Cancel] [Save Alert]                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**
- List of existing alerts with status toggle
- Quick actions (Edit, Delete)
- Form to create new alerts
- Condition builder (log level, threshold, time window)
- Email recipient management
- Template selection
- Frequency throttling

---

## 7. Interaction Patterns

### 7.1 Loading States
- **Initial load (0-300ms)**: Show skeleton screens with shimmer effect
- **Delayed load (>300ms)**: Show spinner + message ("Loading logs...")
- **Pagination**: Loading state on table with subtle overlay
- **Charts**: Skeleton bar charts matching final layout

### 7.2 Error Handling
- **Error messages**: Red background box with error icon + message + retry button
- **Placement**: Below the affected component (form field) or in modal
- **Example**: "Failed to load logs. [Retry]"
- **Network errors**: Show offline message with auto-retry timer

### 7.3 Success Feedback
- **Toast notifications**: Green bg, top-right corner, 4-second auto-dismiss
- **Inline feedback**: Green checkmark inline with saved status
- **Example**: "✓ Alert created successfully"

### 7.4 Real-time Updates
- **Live badge**: Green "● LIVE" indicator in header
- **Auto-refresh**: Tables refresh every 30s in background
- **Pulse animation**: New/updated rows pulse briefly (50ms subtle highlight)
- **Indicator**: "Updated 2 seconds ago" near timestamp

### 7.5 Empty States
- **No logs**: Show icon + message + call-to-action
- **Example**: "No logs found. Adjust filters or check back later. [Clear Filters]"
- **Height**: Fill available space (min 300px)

### 7.6 Confirmation Dialogs
- **Delete actions**: "Are you sure? This cannot be undone. [Cancel] [Delete]"
- **Destructive styling**: Red delete button
- **Focus management**: Cancel button has initial focus

### 7.7 Animations
- **Page transition**: Fade-in (150ms) when navigating between tabs
- **Modal open**: Scrim fade + card scale (200ms)
- **Table row expand**: Slide-down (200ms)
- **Chart data load**: Data bars animate from 0 to final value (400ms, staggered)
- **Button press**: Scale 0.95 on click (100ms)
- **Tooltip**: Fade-in on hover (200ms delay)

---

## 8. Responsive Design

### Mobile-First Approach (375px base)
- **Single column layout** for primary content
- **Filters**: Collapsible bottom sheet or side drawer
- **Table**: Horizontal scroll or card-based layout (one log per card)
- **Charts**: Simplified versions or stacked vertical
- **Buttons**: Full-width (100%) for primary actions
- **Navigation**: Bottom tab bar (max 5 items)

### Tablet Layout (768px+)
- **Two-column grid** for some sections
- **Side-by-side charts** (error trend + distribution)
- **Sticky sidebar** for filters (optional)
- **Table**: Full width with horizontal scroll if needed

### Desktop Layout (1024px+)
- **Three-column or full-width layouts** as needed
- **Side panels** for filters/details
- **Multiple charts** visible simultaneously
- **Keyboard shortcuts** (e.g., "/" for search)

### Safe Areas (Mobile)
- **Top safe area**: 16px gap from notch/status bar
- **Bottom safe area**: 16px gap from gesture bar
- **Side safe areas**: 12px insets from screen edges

---

## 9. Accessibility (WCAG AA)

### Color Contrast
- **Text on background**: 4.5:1 (normal text), 3:1 (large text)
- **Border on background**: 3:1
- **Icon colors**: 3:1 minimum
- **Verification**: Use contrast checker for all color combinations

### Keyboard Navigation
- **Tab order**: Logical, left-to-right, top-to-bottom
- **Focus visible**: 2px blue outline with 2px offset
- **Skip link**: Skip to main content (hidden, shown on focus)
- **Forms**: Labels associated via `<label for="id">`

### Screen Reader Support
- **Alt text**: Meaningful descriptions for charts/images
- **Aria-labels**: Icon-only buttons have aria-label
- **Aria-live**: Loading/error messages use aria-live="polite"
- **Headings**: Semantic hierarchy (h1→h2→h3, no skips)
- **Tables**: `<thead>`, `<tbody>`, `<th>` with scope attribute
- **Forms**: Required fields marked with aria-required="true"

### Dynamic Type & Text Scaling
- **Min font size**: 16px (avoids iOS auto-zoom)
- **Relative units**: Use em/rem for scalability
- **Line height**: 1.5+ for readability
- **No truncation**: Text should wrap, not overflow

### Motion & Transitions
- **Reduced motion**: Respect `prefers-reduced-motion` media query
- **No auto-play**: Videos/animations don't auto-start
- **Interruptible**: User can stop animations by interacting

---

## 10. Performance Targets

### Core Web Vitals
- **LCP** (Largest Contentful Paint): < 2.5s
- **FID** (First Input Delay): < 100ms
- **CLS** (Cumulative Layout Shift): < 0.1

### Optimization Strategies
- **Code splitting**: Route-based dynamic imports
- **Image optimization**: WebP/AVIF with fallbacks
- **Lazy loading**: Defer charts and below-fold content
- **Virtualization**: Virtualize logs table (1000+ rows)
- **Caching**: Cache API responses (5-min TTL)
- **Bundle size**: Keep main bundle <150KB (gzipped)

### Database Query Optimization
- **Pagination**: 20 logs per page (default)
- **Aggregation**: Use hourly tables for charts (not raw logs)
- **Indexing**: Ensure indexes on timestamp, log_level, collector_id
- **Time filters**: Always include time range in queries

---

## 11. Development Patterns

### Component Structure
```
/src
  /components
    /Logs
      LogsDashboard.tsx       (container, state management)
      LogsTable.tsx           (reusable table component)
      LogDetailsModal.tsx      (modal with details)
      FilterPanel.tsx         (filter controls)
    /Analytics
      AnalyticsDashboard.tsx
      ErrorTrendChart.tsx
      LogDistributionChart.tsx
    /Alerts
      AlertsList.tsx
      AlertForm.tsx
    /Common
      Button.tsx
      Modal.tsx
      Card.tsx
      Badge.tsx
      LoadingSpinner.tsx
  /hooks
    useCollectors.ts
    useLogs.ts
    useAlerts.ts
    useAnalytics.ts
  /api
    logClient.ts
    alertClient.ts
  /types
    logs.ts
    alerts.ts
    collectors.ts
  /utils
    formatters.ts
    validators.ts
```

### State Management
- **Zustand** for global state (user, instance selection, alert filters)
- **React hooks** for local component state
- **React Query** (future) for caching API responses

### Form Management
- **React Hook Form** + **Zod** for validation
- **Controlled components** for filters
- **Uncontrolled components** for simple inputs

---

## 12. Implementation Checklist

### Phase 1: Foundation (Logs Dashboard)
- [ ] Setup Tailwind CSS theme configuration
- [ ] Create base Button, Card, Modal, Badge components
- [ ] Implement LogsDashboard layout
- [ ] Build LogsTable with sorting/pagination
- [ ] Create LogDetailsModal
- [ ] Implement FilterPanel (log level, time range)
- [ ] Connect to API (useCollectors hook)

### Phase 2: Analytics & Alerts
- [ ] Build AnalyticsDashboard layout
- [ ] Integrate Recharts (error trend, log distribution charts)
- [ ] Create AlertsList component
- [ ] Build AlertForm (condition builder)
- [ ] Implement email alert configuration flow
- [ ] Add real-time updates (auto-refresh, live badge)

### Phase 3: Polish & Optimization
- [ ] Add loading/error/empty states
- [ ] Implement animations (fade, scale, slide)
- [ ] Test accessibility (WCAG AA compliance)
- [ ] Optimize performance (code splitting, lazy loading)
- [ ] Dark mode support
- [ ] Mobile responsiveness testing
- [ ] Cross-browser testing

### Phase 4: Testing & Documentation
- [ ] Unit tests for components
- [ ] Integration tests for API calls
- [ ] E2E tests for critical flows
- [ ] Storybook for component documentation
- [ ] README with UI guidelines

---

## 13. Color Quick Reference

### Semantic Palette (Tailwind)
```css
/* Backgrounds */
--bg-primary:    #FFFFFF           /* White */
--bg-secondary:  #F9FAFB           /* Gray-50 */
--bg-hover:      #F3F4F6           /* Gray-100 */
--bg-selected:   #EFF6FF           /* Blue-50 */

/* Text */
--text-primary:  #111827           /* Gray-900 */
--text-secondary: #6B7280          /* Gray-500 */
--text-disabled: #D1D5DB           /* Gray-300 */

/* Borders */
--border:        #E5E7EB           /* Gray-200 */
--border-hover:  #D1D5DB           /* Gray-300 */
--border-focus:  #3B82F6           /* Blue-500 */

/* Status */
--success:       #10B981           /* Emerald-500 */
--warning:       #F59E0B           /* Amber-500 */
--error:         #EF4444           /* Red-500 */
--info:          #3B82F6           /* Blue-500 */
```

### Tailwind Config
```javascript
extend: {
  colors: {
    'semantic': {
      success: '#10B981',
      warning: '#F59E0B',
      error: '#EF4444',
      info: '#3B82F6',
    }
  },
  spacing: {
    'xs': '4px',
    'sm': '8px',
    'md': '12px',
    'lg': '16px',
    'xl': '24px',
  },
  fontSize: {
    'display': ['32px', { lineHeight: '1.2' }],
    'h1': ['24px', { lineHeight: '1.2' }],
    'body': ['16px', { lineHeight: '1.5' }],
  },
}
```

---

## 14. Future Enhancements

### Short-term (Next Quarter)
- [ ] Real-time log streaming (WebSocket)
- [ ] Advanced search syntax (Lucene-style queries)
- [ ] Custom dashboard layouts
- [ ] Email alert templates
- [ ] Saved filter sets/presets

### Medium-term (Next 2 Quarters)
- [ ] Dark mode support
- [ ] Mobile app (React Native)
- [ ] API documentation (Swagger)
- [ ] Performance profiling tools
- [ ] Log export (CSV, JSON)
- [ ] Webhook alerts (Slack, PagerDuty)

### Long-term (Next 3+ Quarters)
- [ ] Machine learning anomaly detection
- [ ] Predictive alerting
- [ ] Multi-instance federated search
- [ ] Custom visualization builder
- [ ] API for third-party integrations

---

**Status**: ✅ Design System Complete
**Ready for Implementation**: Yes
**Next Step**: Start Phase 1 (Logs Dashboard components)
