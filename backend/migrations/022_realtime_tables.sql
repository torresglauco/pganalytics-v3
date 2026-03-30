-- Metrics Table (for storing alert-relevant metrics)
CREATE TABLE IF NOT EXISTS metrics (
  id BIGSERIAL PRIMARY KEY,
  instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id) ON DELETE CASCADE,
  metric_name VARCHAR(255) NOT NULL,
  value NUMERIC NOT NULL,
  timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metrics_instance_id ON metrics(instance_id);
CREATE INDEX idx_metrics_metric_name ON metrics(metric_name);
CREATE INDEX idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX idx_metrics_instance_metric_time ON metrics(instance_id, metric_name, timestamp DESC);

-- Notification Channels Table
CREATE TABLE IF NOT EXISTS notification_channels (
  id BIGSERIAL PRIMARY KEY,
  alert_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL, -- email, slack, webhook, pagerduty, sms
  config JSONB NOT NULL, -- Channel-specific configuration
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_channels_alert_id ON notification_channels(alert_id);
CREATE INDEX idx_notification_channels_active ON notification_channels(is_active) WHERE is_active = true;

-- Alert Triggers Table
CREATE TABLE IF NOT EXISTS alert_triggers (
  id BIGSERIAL PRIMARY KEY,
  alert_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
  instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id),
  triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  UNIQUE(alert_id, instance_id, DATE(triggered_at))
);

CREATE INDEX idx_alert_triggers_alert_id ON alert_triggers(alert_id);
CREATE INDEX idx_alert_triggers_instance_id ON alert_triggers(instance_id);
CREATE INDEX idx_alert_triggers_triggered_at ON alert_triggers(triggered_at DESC);
CREATE INDEX idx_alert_triggers_created_at ON alert_triggers(created_at DESC);

-- Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  channel_id BIGINT NOT NULL REFERENCES notification_channels(id),
  alert_trigger_id BIGINT NOT NULL REFERENCES alert_triggers(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, delivered, failed
  retry_count INTEGER DEFAULT 0,
  last_retry_at TIMESTAMP WITH TIME ZONE,
  sent_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel_id ON notifications(channel_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_alert_trigger_id ON notifications(alert_trigger_id);
