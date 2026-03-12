# Task #11: Frontend Updates for Enterprise Auth & Alerts - Implementation Plan

**Status**: In Progress
**Date**: March 5, 2026
**Phase**: Phase 5 - Anomaly Detection & Alerting (v3.5.0)
**Task**: #11 of 12

---

## Objective

Update the frontend to support:
1. Enterprise authentication methods (LDAP, SAML, OAuth, MFA)
2. Alert management dashboard (create, edit, delete alert rules)
3. Real-time alert notifications and incident management
4. Notification channel configuration (Slack, Email, Webhook, PagerDuty, Jira)

---

## Current State Analysis

### Existing Frontend Components

**Auth System**:
- `AuthPage.tsx` - Basic login page with form
- `LoginForm.tsx` - Simple username/password login
- `ChangePasswordForm.tsx` - Password change functionality
- `CreateUserForm.tsx` - User creation form
- `UserManagementTable.tsx` - User management table

**Alert System**:
- `AlertsIncidents.tsx` - Mock alerts dashboard
- `types/alerts.ts` - Alert and incident types defined
- `alertStore.ts` - Zustand store for alert state

### What Needs to Be Built

#### 1. Enterprise Authentication UI
- [ ] Multi-method login selector (LDAP, SAML, OAuth, Local)
- [ ] LDAP login form
- [ ] SAML login handler (redirects to IdP)
- [ ] OAuth provider selector and login flow
- [ ] MFA setup wizard (TOTP, SMS, backup codes)
- [ ] MFA verification during login
- [ ] Session management UI

#### 2. Alert Rules Management
- [ ] Alert rules dashboard with CRUD operations
- [ ] Rule creation form (threshold, anomaly, change, composite)
- [ ] Rule editing interface
- [ ] Rule testing modal
- [ ] Rule deletion with confirmation
- [ ] Bulk rule actions (enable/disable/delete)

#### 3. Notification Channels
- [ ] Channel configuration form (Slack, Email, Webhook, PagerDuty, Jira)
- [ ] Channel management table
- [ ] Test channel connectivity button
- [ ] Delivery history view
- [ ] Channel status indicator

#### 4. Alert Dashboard Enhancements
- [ ] Real-time alert updates (WebSocket or polling)
- [ ] Alert filtering and search
- [ ] Bulk alert actions (acknowledge/resolve)
- [ ] Alert detail modal with history
- [ ] Incident grouping and correlation
- [ ] Alert metrics and KPIs

#### 5. UI Components & Utilities
- [ ] Auth context/provider for session management
- [ ] Protected route component
- [ ] API client functions for backend integration
- [ ] WebSocket hook for real-time updates
- [ ] Form builders for rule/channel configuration

---

## Implementation Plan

### Phase 1: Enterprise Authentication (Priority 1)

**Duration**: 2-3 hours
**Components to Create**:
1. `EnterpriseAuthPage.tsx` (450 lines)
   - Multi-method login selector
   - Route to appropriate auth method
   - Error handling and messaging

2. `LDAPLoginForm.tsx` (280 lines)
   - Username/password fields specific to LDAP
   - LDAP server address display
   - Connection error handling

3. `SAMLLoginHandler.tsx` (150 lines)
   - Redirect to SAML IdP
   - Handle assertion callback
   - Session establishment

4. `OAuthProviderSelector.tsx` (200 lines)
   - Provider button list (Google, Azure AD, GitHub)
   - Custom OIDC provider support
   - OAuth flow initialization

5. `MFASetupWizard.tsx` (500 lines)
   - TOTP setup with QR code display
   - SMS setup with verification
   - Backup codes generation and download
   - Multi-step wizard flow

6. `MFAVerificationForm.tsx` (250 lines)
   - TOTP code entry
   - SMS code entry
   - Backup code fallback
   - Remember device option

7. `contexts/AuthContext.tsx` (300 lines)
   - Session state management
   - User profile with auth method
   - Login/logout handlers
   - Protected route wrapper

8. `hooks/useAuth.ts` (100 lines)
   - Auth context hook
   - Session refresh logic
   - Token management

**Files Modified**:
- `AuthPage.tsx` - Replace with enterprise version
- `types/index.ts` - Add auth types (AuthMethod, User, Session)
- `store/authStore.ts` - Add MFA state, session info

**Success Criteria**:
- ✅ Login with username/password works
- ✅ LDAP integration displays properly
- ✅ SAML redirects to IdP
- ✅ OAuth provider selection works
- ✅ MFA setup wizard functions
- ✅ MFA verification during login works
- ✅ Session persists across navigation
- ✅ Protected routes redirect unauthenticated users

---

### Phase 2: Alert Rules Management (Priority 2)

**Duration**: 3-4 hours
**Components to Create**:
1. `AlertRulesPage.tsx` (600 lines)
   - Rules management dashboard
   - Create, edit, delete, test, enable/disable
   - Filtering and sorting
   - Bulk actions

2. `AlertRuleForm.tsx` (700 lines)
   - Four rule type variants (threshold, anomaly, change, composite)
   - Condition builder with visual operators
   - Metric/database selector
   - Threshold/value input fields
   - Severity level selection
   - Testing interface

3. `AlertRuleDetailModal.tsx` (350 lines)
   - Rule display with all details
   - Edit mode toggle
   - Test rule interface
   - Rule history/evaluation log
   - Delete confirmation

4. `CompositeConditionBuilder.tsx` (400 lines)
   - Add/remove sub-conditions
   - AND/OR operator selection
   - Nested condition support
   - Visual tree representation

5. `RuleTestingModal.tsx` (300 lines)
   - Test rule with sample data
   - Display predicted firing behavior
   - Show evaluation timeline
   - Validate condition logic

6. `AlertRulesTable.tsx` (350 lines)
   - Sortable columns (name, type, severity, status)
   - Status indicators (enabled/disabled)
   - Last evaluation timestamp
   - Quick actions (edit, test, delete)
   - Bulk selection and actions

7. `types/alertRules.ts` (250 lines)
   - AlertRule interface
   - RuleCondition union type
   - RuleType enum
   - API request/response types

8. `api/alertRulesApi.ts` (200 lines)
   - CRUD endpoints
   - Test rule endpoint
   - Bulk operations
   - Error handling

**Files Modified**:
- `AlertsIncidents.tsx` - Add navigation to rules management
- `SettingsAdmin.tsx` - Link to alert rules in settings

**Success Criteria**:
- ✅ Create threshold rule with operator and value
- ✅ Create anomaly rule with severity levels
- ✅ Create change rule with percentage threshold
- ✅ Create composite rule with AND/OR logic
- ✅ Test rule shows correct firing behavior
- ✅ Edit rule updates properly
- ✅ Delete rule with confirmation
- ✅ Enable/disable rule toggles status
- ✅ Bulk actions work on multiple rules

---

### Phase 3: Notification Channels Configuration (Priority 2)

**Duration**: 2-3 hours
**Components to Create**:
1. `NotificationChannelsPage.tsx` (500 lines)
   - Channel management dashboard
   - Create new channel button
   - Channels table/card view
   - Quick actions and bulk operations

2. `ChannelConfigForm.tsx` (600 lines)
   - Dynamic form based on channel type
   - Slack: webhook URL input with validation
   - Email: SMTP host, port, from address, recipients
   - Webhook: URL, method, headers, auth
   - PagerDuty: integration key, severity mapping
   - Jira: URL, token, project key, issue type

3. `ChannelTestModal.tsx` (250 lines)
   - Send test notification
   - Display delivery result
   - Show error messages if failed
   - Retry button

4. `DeliveryHistoryTable.tsx` (300 lines)
   - List recent deliveries
   - Status per message (success/failed/retrying)
   - Retry count
   - Last attempt timestamp
   - Error details

5. `ChannelStatusIndicator.tsx` (150 lines)
   - Health status display (healthy/degraded/unhealthy)
   - Recent failure count
   - Success rate percentage
   - Tooltip with details

6. `types/channels.ts` (200 lines)
   - Channel type definitions
   - ChannelConfig interface per type
   - Delivery result types
   - API request/response types

7. `api/channelsApi.ts` (200 lines)
   - Create/read/update/delete channels
   - Test channel endpoint
   - Get delivery history
   - Get channel status

**Files Modified**:
- `SettingsAdmin.tsx` - Add channels management section

**Success Criteria**:
- ✅ Create Slack channel with webhook validation
- ✅ Create Email channel with SMTP config
- ✅ Create Webhook with custom headers
- ✅ Create PagerDuty with API key
- ✅ Create Jira channel with credentials
- ✅ Test channel sends notification
- ✅ View delivery history
- ✅ Channel status shows health
- ✅ Edit channel configuration
- ✅ Delete channel with confirmation

---

### Phase 4: Alert Dashboard Enhancements (Priority 3)

**Duration**: 3-4 hours
**Components to Create**:
1. `AlertDashboardEnhanced.tsx` (700 lines)
   - Alert metrics (critical, high, warning, info)
   - Trends chart (alerts over time)
   - Top alert types
   - Top affected databases
   - Filter and search interface

2. `AlertDetailModal.tsx` (400 lines)
   - Alert full details and context
   - Alert history timeline
   - Related anomalies
   - Suggested remediation
   - Acknowledge/resolve/mute buttons
   - Notes field

3. `AlertFilterPanel.tsx` (350 lines)
   - Filter by severity (critical, high, warning, info)
   - Filter by status (active, acknowledged, resolved)
   - Filter by alert type
   - Filter by database/collector
   - Date range picker
   - Save filter presets

4. `IncidentGroupView.tsx` (500 lines)
   - Group related alerts into incidents
   - Incident detail with all associated alerts
   - Root cause hypothesis
   - Confidence score
   - Suggested actions
   - Timeline of events

5. `RealTimeUpdatesHook.ts` (200 lines)
   - WebSocket hook for alert updates
   - Fallback polling if WebSocket unavailable
   - Reconnection logic
   - Message handling

6. `AlertMetricsCards.tsx` (250 lines)
   - Critical alerts count
   - High alerts count
   - Average resolution time
   - SLA compliance percentage

7. `api/alertsApi.ts` (200 lines)
   - Get alerts with filtering
   - Acknowledge alert
   - Resolve alert
   - Mute alert
   - Add notes to alert
   - Get incident details

**Files Modified**:
- `AlertsIncidents.tsx` - Replace with enhanced version

**Success Criteria**:
- ✅ Alerts dashboard shows real-time updates
- ✅ Filter alerts by multiple criteria
- ✅ View alert detail with history
- ✅ Acknowledge/resolve/mute alert
- ✅ Add notes to alert
- ✅ Group related alerts as incidents
- ✅ View incident root cause and actions
- ✅ Dashboard metrics update in real-time
- ✅ Search alerts by keyword
- ✅ Save and restore filter presets

---

### Phase 5: UI Polish & Integration (Priority 3)

**Duration**: 2-3 hours
**Components to Create**:
1. `ProtectedRoute.tsx` (100 lines)
   - Route wrapper for auth validation
   - Redirect to login if unauthorized
   - Loading state while checking auth

2. `AuthenticationStatus.tsx` (200 lines)
   - Current auth method display
   - MFA status indicator
   - Session expiration warning
   - Re-login prompt

3. `AlertNotificationBell.tsx` (300 lines)
   - Notification badge with unread count
   - Dropdown menu with recent alerts
   - Quick access to acknowledge/resolve
   - Bell icon with animation on new alerts

4. `EmptyStates.tsx` (150 lines)
   - No alerts empty state
   - No rules empty state
   - No channels empty state
   - Helpful action buttons

**Files to Create**:
- `layouts/WithAlertNotifications.tsx` (150 lines)
  - Wrapper component that adds alert bell
  - Real-time updates subscription

**Files Modified**:
- `App.tsx` - Use ProtectedRoute for pages
- `components/index.ts` - Export new components
- `types/index.ts` - Consolidate types

**Success Criteria**:
- ✅ Unauthorized users redirected to login
- ✅ All routes properly protected
- ✅ Alert notification bell shows updates
- ✅ Empty states guide users to create resources
- ✅ UI is polished and professional
- ✅ All pages integrated properly

---

## File Structure

```
frontend/src/
├── pages/
│   ├── AuthPage.tsx (UPDATED)
│   ├── AlertRulesPage.tsx (NEW)
│   ├── NotificationChannelsPage.tsx (NEW)
│   ├── AlertsIncidents.tsx (UPDATED)
│   └── SettingsAdmin.tsx (MODIFIED to add new sections)
│
├── components/
│   ├── EnterpriseAuthPage.tsx (NEW)
│   ├── LDAPLoginForm.tsx (NEW)
│   ├── SAMLLoginHandler.tsx (NEW)
│   ├── OAuthProviderSelector.tsx (NEW)
│   ├── MFASetupWizard.tsx (NEW)
│   ├── MFAVerificationForm.tsx (NEW)
│   ├── ProtectedRoute.tsx (NEW)
│   ├── AuthenticationStatus.tsx (NEW)
│   ├── AlertNotificationBell.tsx (NEW)
│   ├── AlertRuleForm.tsx (NEW)
│   ├── AlertRuleDetailModal.tsx (NEW)
│   ├── CompositeConditionBuilder.tsx (NEW)
│   ├── RuleTestingModal.tsx (NEW)
│   ├── AlertRulesTable.tsx (NEW)
│   ├── ChannelConfigForm.tsx (NEW)
│   ├── ChannelTestModal.tsx (NEW)
│   ├── DeliveryHistoryTable.tsx (NEW)
│   ├── ChannelStatusIndicator.tsx (NEW)
│   ├── AlertDetailModal.tsx (NEW)
│   ├── AlertFilterPanel.tsx (NEW)
│   ├── IncidentGroupView.tsx (NEW)
│   ├── AlertMetricsCards.tsx (NEW)
│   ├── EmptyStates.tsx (NEW)
│   └── LoginForm.tsx (MODIFIED)
│
├── contexts/
│   └── AuthContext.tsx (NEW)
│
├── hooks/
│   ├── useAuth.ts (NEW)
│   └── useRealTimeAlerts.ts (NEW)
│
├── api/
│   ├── alertRulesApi.ts (NEW)
│   ├── channelsApi.ts (NEW)
│   ├── alertsApi.ts (NEW)
│   └── authApi.ts (UPDATED)
│
├── types/
│   ├── index.ts (UPDATED - add auth types)
│   ├── alerts.ts (UPDATED - expand alert types)
│   ├── alertRules.ts (NEW)
│   └── channels.ts (NEW)
│
├── store/
│   ├── authStore.ts (NEW/UPDATED)
│   ├── alertRulesStore.ts (NEW)
│   ├── channelsStore.ts (NEW)
│   └── alertStore.ts (UPDATED)
│
├── layouts/
│   └── WithAlertNotifications.tsx (NEW)
│
└── App.tsx (UPDATED - add new routes and protection)
```

---

## Estimated Effort

| Phase | Duration | Components | Lines |
|-------|----------|-----------|-------|
| Phase 1: Auth | 2-3h | 8 | 2,200 |
| Phase 2: Rules | 3-4h | 8 | 2,600 |
| Phase 3: Channels | 2-3h | 7 | 2,000 |
| Phase 4: Dashboard | 3-4h | 7 | 2,700 |
| Phase 5: Polish | 2-3h | 4 | 600 |
| **Total** | **12-17h** | **34** | **10,100** |

---

## Implementation Order

1. **Week 1 (Phase 1)**: Enterprise Authentication UI
   - Create auth context and types
   - Build enterprise auth page and forms
   - Integrate with backend auth endpoints
   - Test MFA flows

2. **Week 2 (Phase 2-3)**: Rules & Channels Management
   - Build alert rules dashboard and forms
   - Create notification channels configuration
   - Integrate with backend APIs
   - Add testing capabilities

3. **Week 3 (Phase 4-5)**: Dashboard & Polish
   - Enhance alert dashboard with real-time updates
   - Add UI polish and empty states
   - Integrate all components
   - Final testing and refinements

---

## Key Integration Points

### Backend API Endpoints Required
```
POST /api/v1/auth/ldap/login
POST /api/v1/auth/saml/initiate
POST /api/v1/auth/saml/callback
GET  /api/v1/auth/oauth/providers
POST /api/v1/auth/oauth/:provider/login
POST /api/v1/auth/oauth/callback
POST /api/v1/auth/mfa/setup
POST /api/v1/auth/mfa/verify

POST   /api/v1/alert-rules
GET    /api/v1/alert-rules
GET    /api/v1/alert-rules/:id
PUT    /api/v1/alert-rules/:id
DELETE /api/v1/alert-rules/:id
POST   /api/v1/alert-rules/:id/test

POST   /api/v1/notification-channels
GET    /api/v1/notification-channels
GET    /api/v1/notification-channels/:id
PUT    /api/v1/notification-channels/:id
DELETE /api/v1/notification-channels/:id
POST   /api/v1/notification-channels/:id/test
GET    /api/v1/notification-channels/:id/history

GET    /api/v1/alerts
GET    /api/v1/alerts/:id
POST   /api/v1/alerts/:id/acknowledge
POST   /api/v1/alerts/:id/resolve
POST   /api/v1/alerts/:id/mute
POST   /api/v1/alerts/:id/notes
```

### State Management
- Auth state: Current user, session, auth method, MFA status
- Rules state: Rules list, selected rule, rule filters
- Channels state: Channels list, delivery history
- Alerts state: Alerts list, filters, selected alert, real-time updates

### WebSocket Events
- `alert:new` - New alert created
- `alert:updated` - Alert status changed
- `alert:resolved` - Alert resolved
- `rule:created` - New rule created
- `rule:updated` - Rule updated
- `channel:status` - Channel status changed

---

## Success Criteria for Task #11

- ✅ All 34 components created and functional
- ✅ 10,100+ lines of frontend code added
- ✅ Enterprise auth methods fully integrated (LDAP, SAML, OAuth, MFA)
- ✅ Alert rules management CRUD working
- ✅ Notification channels configuration complete
- ✅ Real-time alert dashboard operational
- ✅ All routes properly protected
- ✅ UI polished and professional
- ✅ Integrated with Phase 5 backend APIs
- ✅ No console errors or warnings
- ✅ Responsive design on mobile/tablet/desktop
- ✅ Accessibility standards met (WCAG 2.1 AA)

---

## Next Steps

1. Start Phase 1: Enterprise Authentication UI
2. Create auth context and types
3. Build login page with multi-method selector
4. Implement LDAP, SAML, OAuth forms
5. Build MFA setup and verification
6. Integrate with backend auth endpoints
7. Test complete authentication flow
8. Commit and push changes
9. Move to Phase 2: Rules Management

---

**Plan Created**: March 5, 2026
**Status**: Ready to implement
**Priority**: HIGH (Frontend completion for v3.5.0)
