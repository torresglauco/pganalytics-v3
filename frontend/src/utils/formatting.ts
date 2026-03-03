export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + ' ' + sizes[i];
}

export function formatPercentage(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatDuration(milliseconds: number): string {
  if (milliseconds < 1) return `${(milliseconds * 1000).toFixed(0)}µs`;
  if (milliseconds < 1000) return `${milliseconds.toFixed(2)}ms`;
  if (milliseconds < 60000) return `${(milliseconds / 1000).toFixed(2)}s`;
  return `${(milliseconds / 60000).toFixed(2)}m`;
}

export function formatNumber(num: number, decimals = 0): string {
  return num.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

export function getHealthColor(score: number): string {
  if (score >= 80) return 'text-pg-success';
  if (score >= 60) return 'text-pg-warning';
  return 'text-pg-danger';
}

export function getHealthBgColor(score: number): string {
  if (score >= 80) return 'bg-pg-success/10';
  if (score >= 60) return 'bg-pg-warning/10';
  return 'bg-pg-danger/10';
}

export function formatDateTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function getRelativeTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return formatDateTime(d);
}
