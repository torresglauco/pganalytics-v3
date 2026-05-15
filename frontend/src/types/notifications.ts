/**
 * Notification Channels Type Definitions
 * Comprehensive types for notification channel configuration and management
 */

/**
 * Notification channel types
 */
export type ChannelType = 'slack' | 'email' | 'webhook' | 'pagerduty' | 'opsgenie' | 'jira';

/**
 * Channel enabled/disabled state
 */
export type ChannelStatus = 'active' | 'inactive' | 'testing' | 'error';

/**
 * Notification delivery status
 */
export type DeliveryStatus = 'pending' | 'sent' | 'failed' | 'bounced';

/**
 * Base notification channel configuration
 */
export interface NotificationChannel {
  id: string;
  name: string;
  description?: string;
  type: ChannelType;
  status: ChannelStatus;
  is_default?: boolean;

  // Configuration (type-specific)
  config: SlackConfig | EmailConfig | WebhookConfig | PagerDutyConfig | OpsGenieConfig | JiraConfig;

  // Metadata
  created_at: string;
  updated_at: string;
  created_by: string;
  last_test_at?: string;
  last_test_status?: boolean;
  last_error?: string;

  // Stats
  total_sent?: number;
  success_count?: number;
  failure_count?: number;
}

/**
 * Slack channel configuration
 */
export interface SlackConfig {
  webhook_url: string; // Slack webhook URL
  channel?: string; // Channel to post to (optional, uses webhook default)
  username?: string; // Bot username
  icon_emoji?: string; // Bot emoji
  mentions?: string; // User/group mentions on critical alerts
  thread_replies?: boolean; // Thread replies for alert updates
}

/**
 * Email channel configuration
 */
export interface EmailConfig {
  smtp_server: string;
  smtp_port: number;
  smtp_username?: string;
  smtp_password?: string;
  smtp_use_tls: boolean;
  from_address: string;
  from_name?: string;
  recipients: string[]; // List of recipient emails
  cc?: string[];
  bcc?: string[];
  html_template?: string; // Custom HTML template
}

/**
 * Generic webhook channel configuration
 */
export interface WebhookConfig {
  url: string;
  method: 'POST' | 'PUT' | 'PATCH';
  headers?: Record<string, string>;
  auth_type?: 'basic' | 'bearer' | 'api_key' | 'oauth2' | 'none';
  auth_credentials?: {
    username?: string;
    password?: string;
    token?: string;
    api_key_header?: string;
    api_key_value?: string;
  };
  payload_template?: string; // JSON template for request body
  retry_enabled?: boolean;
  retry_max_attempts?: number;
}

/**
 * PagerDuty channel configuration
 */
export interface PagerDutyConfig {
  integration_key: string; // Integration key for service
  service_id?: string;
  escalation_policy_id?: string;
  urgency?: 'low' | 'high';
  custom_details?: Record<string, string>;
}

/**
 * OpsGenie channel configuration
 */
export interface OpsGenieConfig {
  api_key: string; // OpsGenie API key
  region?: 'us' | 'eu'; // API region (default: us)
  team_id?: string; // Team ID for routing
  priority_mapping?: {
    low: 'P4';
    medium: 'P3';
    high: 'P2';
    critical: 'P1';
  };
  tags?: string[]; // Additional tags for alerts
}

/**
 * Jira channel configuration
 */
export interface JiraConfig {
  base_url: string;
  project_key: string;
  issue_type?: string; // e.g., "Bug", "Incident"
  username: string;
  api_token: string;
  custom_fields?: Record<string, string>;
  priority_mapping?: {
    low: string;
    medium: string;
    high: string;
    critical: string;
  };
  auto_close?: boolean; // Auto-close issue when alert resolves
}

/**
 * Create notification channel request
 */
export interface CreateChannelRequest {
  name: string;
  description?: string;
  type: ChannelType;
  config:
    | SlackConfig
    | EmailConfig
    | WebhookConfig
    | PagerDutyConfig
    | OpsGenieConfig
    | JiraConfig;
}

/**
 * Update notification channel request
 */
export interface UpdateChannelRequest extends Partial<CreateChannelRequest> {
  id: string;
}

/**
 * Test channel result
 */
export interface TestChannelResult {
  channel_id: string;
  success: boolean;
  message: string;
  response_time_ms: number;
  status_code?: number;
  error?: string;
}

/**
 * Notification delivery record
 */
export interface NotificationDelivery {
  id: string;
  channel_id: string;
  alert_id: string;
  rule_id: string;
  status: DeliveryStatus;
  sent_at: string;
  delivered_at?: string;
  bounced_at?: string;
  response_time_ms: number;
  error_message?: string;
  retry_count: number;
}

/**
 * Delivery history statistics
 */
export interface DeliveryStats {
  channel_id: string;
  total: number;
  sent: number;
  failed: number;
  bounced: number;
  avg_response_time_ms: number;
  success_rate: number;
  last_failure_at?: string;
  last_failure_reason?: string;
}

/**
 * Notification template
 */
export interface NotificationTemplate {
  id: string;
  name: string;
  channel_type: ChannelType;
  template: string; // Handlebars/Mustache template
  variables: string[]; // Available variables
  preview_example?: Record<string, any>;
}

/**
 * Bulk channel action request
 */
export interface BulkChannelAction {
  action: 'enable' | 'disable' | 'delete' | 'set_default';
  channel_ids: string[];
}
