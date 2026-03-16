import React, { useState, useMemo } from 'react';
import {
  Plus,
  Trash2,
  Eye,
  EyeOff,
  Copy,
  Check,
  AlertTriangle,
  Bell,
  User,
  Lock,
  Settings as SettingsIcon,
  RotateCw,
  X,
} from 'lucide-react';
import { PageWrapper } from '../components/common/PageWrapper';
import { StatusBadge } from '../components/cards/StatusBadge';
import { DataTable, Column } from '../components/tables/DataTable';
import { formatDateTime } from '../utils/formatting';
import { useUsers, type AdminUser } from '../hooks/useUsers';

interface ApiToken {
  id: string;
  name: string;
  token: string;
  created_at: Date;
  last_used: Date | null;
  expires_at: Date | null;
}

interface NotificationChannel {
  id: string;
  type: 'email' | 'slack' | 'pagerduty' | 'webhook';
  name: string;
  config: {
    email?: string;
    webhook_url?: string;
    slack_channel?: string;
    pagerduty_key?: string;
  };
  enabled: boolean;
  test_status?: 'pending' | 'success' | 'failed';
}


const mockTokens: ApiToken[] = [
  {
    id: '1',
    name: 'Production Monitor',
    token: 'sk_prod_1234567890abcdef1234567890abc',
    created_at: new Date(Date.now() - 30 * 24 * 3600000),
    last_used: new Date(Date.now() - 2 * 3600000),
    expires_at: new Date(Date.now() + 365 * 24 * 3600000),
  },
  {
    id: '2',
    name: 'Staging Backup',
    token: 'sk_staging_abcdef1234567890abcdef1234',
    created_at: new Date(Date.now() - 60 * 24 * 3600000),
    last_used: new Date(Date.now() - 15 * 60000),
    expires_at: new Date(Date.now() + 180 * 24 * 3600000),
  },
  {
    id: '3',
    name: 'CI/CD Pipeline',
    token: 'sk_ci_fedcba0987654321fedcba098765',
    created_at: new Date(Date.now() - 7 * 24 * 3600000),
    last_used: new Date(Date.now() - 30 * 60000),
    expires_at: null,
  },
];

const mockNotifications: NotificationChannel[] = [
  {
    id: '1',
    type: 'email',
    name: 'Admin Alerts',
    config: { email: 'admin@company.com' },
    enabled: true,
    test_status: 'success',
  },
  {
    id: '2',
    type: 'slack',
    name: 'Database Team',
    config: { slack_channel: '#database-alerts' },
    enabled: true,
    test_status: 'success',
  },
  {
    id: '3',
    type: 'pagerduty',
    name: 'On-Call Escalation',
    config: { pagerduty_key: 'REDACTED' },
    enabled: true,
    test_status: 'pending',
  },
  {
    id: '4',
    type: 'webhook',
    name: 'Custom Integration',
    config: { webhook_url: 'https://api.company.com/webhooks/pganalytics' },
    enabled: false,
    test_status: 'failed',
  },
];

export const SettingsAdmin: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'users' | 'tokens' | 'notifications'>('users');
  const [showNewUserForm, setShowNewUserForm] = useState(false);
  const [showNewTokenForm, setShowNewTokenForm] = useState(false);
  const [visibleTokens, setVisibleTokens] = useState<Set<string>>(new Set());
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [resetPasswordConfirm, setResetPasswordConfirm] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // Use the useUsers hook
  const { users, loading, error, fetchUsers, createUser, deleteUser, resetPassword } = useUsers();

  const [newUser, setNewUser] = useState({
    email: '',
    name: '',
    username: '',
    password: '',
    role: 'viewer' as 'admin' | 'user' | 'viewer',
  });

  const [newToken, setNewToken] = useState({
    name: '',
  });

  // Clear success/error messages after 5 seconds
  React.useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [successMessage]);

  React.useEffect(() => {
    if (errorMessage) {
      const timer = setTimeout(() => setErrorMessage(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [errorMessage]);

  const toggleTokenVisibility = (id: string) => {
    const newSet = new Set(visibleTokens);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setVisibleTokens(newSet);
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleAddUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUser.email || !newUser.name) {
      setErrorMessage('Email and name are required');
      return;
    }

    setSubmitting(true);
    try {
      await createUser({
        email: newUser.email,
        name: newUser.name,
        username: newUser.username || newUser.email,
        password: newUser.password || 'TempPassword123!',
        role: newUser.role,
      });
      setSuccessMessage('User created successfully');
      setShowNewUserForm(false);
      setNewUser({ email: '', name: '', username: '', password: '', role: 'viewer' as 'admin' | 'user' | 'viewer' });
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to create user');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteUser = async (userId: string | number) => {
    setSubmitting(true);
    try {
      await deleteUser(userId);
      setSuccessMessage('User deleted successfully');
      setDeleteConfirm(null);
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to delete user');
    } finally {
      setSubmitting(false);
    }
  };

  const handleResetPassword = async (userId: string | number) => {
    setSubmitting(true);
    try {
      const response = await resetPassword(userId);
      setSuccessMessage(`Password reset. Temporary password: ${response.temp_password}`);
      setResetPasswordConfirm(null);
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to reset password');
    } finally {
      setSubmitting(false);
    }
  };

  const handleAddToken = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Adding token:', newToken);
    setShowNewTokenForm(false);
    setNewToken({ name: '' });
  };

  const userColumns: Column<AdminUser>[] = [
    {
      key: 'name',
      label: 'User',
      sortable: true,
      render: (value, row) => (
        <div className="space-y-1">
          <div className="font-medium text-pg-dark">{String(value)}</div>
          <div className="text-xs text-pg-slate">{row.email}</div>
        </div>
      ),
    },
    {
      key: 'role',
      label: 'Role',
      width: '80px',
      render: (value) => {
        const roleColors: Record<string, string> = {
          admin: 'bg-pg-danger/10 text-pg-danger',
          user: 'bg-pg-warning/10 text-pg-warning',
          viewer: 'bg-pg-blue/10 text-pg-blue',
        };
        return (
          <span className={`text-xs font-medium px-2 py-1 rounded ${roleColors[String(value)]}`}>
            {String(value).toUpperCase()}
          </span>
        );
      },
    },
    {
      key: 'status',
      label: 'Status',
      width: '80px',
      render: (value) => (
        <StatusBadge
          status={value === 'active' ? 'success' : 'warning'}
          label={String(value).toUpperCase()}
          size="sm"
        />
      ),
    },
    {
      key: 'last_login',
      label: 'Last Login',
      width: '140px',
      render: (value) => <span className="text-sm text-pg-slate">{formatDateTime(value as Date)}</span>,
    },
  ];

  const tokenColumns: Column<ApiToken>[] = [
    {
      key: 'name',
      label: 'Token Name',
      sortable: true,
      render: (value) => <span className="font-medium text-pg-dark">{String(value)}</span>,
    },
    {
      key: 'created_at',
      label: 'Created',
      width: '120px',
      render: (value) => <span className="text-sm text-pg-slate">{formatDateTime(value as Date)}</span>,
    },
    {
      key: 'last_used',
      label: 'Last Used',
      width: '120px',
      render: (value) => (
        <span className="text-sm text-pg-slate">
          {value ? formatDateTime(value as Date) : 'Never'}
        </span>
      ),
    },
    {
      key: 'expires_at',
      label: 'Expires',
      width: '120px',
      render: (value) => (
        <span className="text-sm text-pg-slate">
          {value ? formatDateTime(value as Date) : 'Never'}
        </span>
      ),
    },
  ];

  return (
    <PageWrapper
      title="Settings & Administration"
      description="Manage users, API tokens, and notification channels"
    >
      {/* Success/Error Messages */}
      {successMessage && (
        <div className="mb-6 p-4 bg-pg-success/10 border border-pg-success/30 rounded-lg flex items-start justify-between">
          <p className="text-pg-success font-medium">{successMessage}</p>
          <button
            onClick={() => setSuccessMessage(null)}
            className="text-pg-success hover:text-pg-success/80"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}
      {(errorMessage || error) && (
        <div className="mb-6 p-4 bg-pg-danger/10 border border-pg-danger/30 rounded-lg flex items-start justify-between">
          <p className="text-pg-danger font-medium">{errorMessage || error?.message}</p>
          <button
            onClick={() => {
              setErrorMessage(null);
            }}
            className="text-pg-danger hover:text-pg-danger/80"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Tabs */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="flex border-b border-pg-slate/20">
          <button
            onClick={() => setActiveTab('users')}
            className={`flex-1 px-6 py-4 font-medium text-center transition-colors ${
              activeTab === 'users'
                ? 'border-b-2 border-pg-blue text-pg-blue'
                : 'text-pg-slate hover:text-pg-dark'
            }`}
          >
            <div className="flex items-center justify-center gap-2">
              <User className="w-4 h-4" />
              Users & Roles
            </div>
          </button>
          <button
            onClick={() => setActiveTab('tokens')}
            className={`flex-1 px-6 py-4 font-medium text-center transition-colors ${
              activeTab === 'tokens'
                ? 'border-b-2 border-pg-blue text-pg-blue'
                : 'text-pg-slate hover:text-pg-dark'
            }`}
          >
            <div className="flex items-center justify-center gap-2">
              <Lock className="w-4 h-4" />
              API Tokens
            </div>
          </button>
          <button
            onClick={() => setActiveTab('notifications')}
            className={`flex-1 px-6 py-4 font-medium text-center transition-colors ${
              activeTab === 'notifications'
                ? 'border-b-2 border-pg-blue text-pg-blue'
                : 'text-pg-slate hover:text-pg-dark'
            }`}
          >
            <div className="flex items-center justify-center gap-2">
              <Bell className="w-4 h-4" />
              Notifications
            </div>
          </button>
        </div>
      </div>

      {/* Users Tab */}
      {activeTab === 'users' && (
        <div className="space-y-6">
          {/* New User Form */}
          {showNewUserForm && (
            <div className="bg-white rounded-lg shadow p-6 border-l-4 border-pg-blue">
              <h3 className="text-lg font-semibold text-pg-dark mb-4">Add New User</h3>
              <form onSubmit={handleAddUser}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Full Name</label>
                    <input
                      type="text"
                      placeholder="John Doe"
                      value={newUser.name}
                      onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Email</label>
                    <input
                      type="email"
                      placeholder="john@company.com"
                      value={newUser.email}
                      onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Username</label>
                    <input
                      type="text"
                      placeholder="johndoe"
                      value={newUser.username}
                      onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Initial Password</label>
                    <input
                      type="password"
                      placeholder="Leave empty for auto-generated"
                      value={newUser.password}
                      onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Role</label>
                    <select
                      value={newUser.role}
                      onChange={(e) =>
                        setNewUser({
                          ...newUser,
                          role: e.target.value as 'admin' | 'user' | 'viewer',
                        })
                      }
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                    >
                      <option value="viewer">Viewer (Read-only access)</option>
                      <option value="user">User (Standard access)</option>
                      <option value="admin">Admin (Full access)</option>
                    </select>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 disabled:bg-pg-slate/50 transition-colors text-sm font-medium"
                  >
                    {submitting ? 'Creating...' : 'Add User'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowNewUserForm(false)}
                    className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Users Table */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-pg-dark">Users</h3>
              <div className="flex gap-2">
                <button
                  onClick={() => fetchUsers()}
                  disabled={loading}
                  className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium disabled:opacity-50"
                  title="Refresh users list"
                >
                  <RotateCw className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setShowNewUserForm(!showNewUserForm)}
                  className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
                >
                  <Plus className="w-4 h-4" />
                  Add User
                </button>
              </div>
            </div>
            {loading ? (
              <div className="text-center py-8">
                <p className="text-pg-slate">Loading users...</p>
              </div>
            ) : (
              <>
                <DataTable
                  columns={userColumns}
                  data={users.map((user) => ({
                    ...user,
                    status: user.is_active !== false ? 'active' : 'inactive',
                    last_login: user.last_login ? new Date(user.last_login) : null,
                    created_at: user.created_at ? new Date(user.created_at) : new Date(),
                  }))}
                  searchable={true}
                  emptyMessage="No users found"
                />
                {/* User Actions */}
                <div className="mt-6 space-y-2">
                  <h4 className="text-sm font-semibold text-pg-dark mb-3">User Actions</h4>
                  {users.length === 0 ? (
                    <p className="text-sm text-pg-slate">No users to manage</p>
                  ) : (
                    <div className="space-y-2 max-h-64 overflow-y-auto">
                      {users.map((user) => (
                        <div key={user.id} className="flex items-center justify-between p-3 bg-pg-slate/5 rounded-lg">
                          <div>
                            <p className="text-sm font-medium text-pg-dark">{user.name || user.username}</p>
                            <p className="text-xs text-pg-slate">{user.email}</p>
                          </div>
                          <div className="flex gap-2">
                            {resetPasswordConfirm === String(user.id) ? (
                              <div className="flex items-center gap-2">
                                <span className="text-xs text-pg-warning font-medium">Confirm?</span>
                                <button
                                  onClick={() => handleResetPassword(user.id)}
                                  disabled={submitting}
                                  className="px-2 py-1 bg-pg-warning text-white rounded text-xs hover:bg-pg-warning/90 disabled:opacity-50"
                                >
                                  {submitting ? '...' : 'Yes'}
                                </button>
                                <button
                                  onClick={() => setResetPasswordConfirm(null)}
                                  className="px-2 py-1 border border-pg-slate/20 text-pg-dark rounded text-xs hover:bg-pg-slate/5"
                                >
                                  No
                                </button>
                              </div>
                            ) : (
                              <button
                                onClick={() => setResetPasswordConfirm(String(user.id))}
                                className="px-3 py-1 border border-pg-slate/20 text-pg-slate rounded text-xs hover:bg-pg-slate/5 transition-colors flex items-center gap-1"
                                title="Reset password to temporary value"
                              >
                                <RotateCw className="w-3 h-3" />
                                Reset Password
                              </button>
                            )}
                            {deleteConfirm === String(user.id) ? (
                              <div className="flex items-center gap-2">
                                <span className="text-xs text-pg-danger font-medium">Delete?</span>
                                <button
                                  onClick={() => handleDeleteUser(user.id)}
                                  disabled={submitting}
                                  className="px-2 py-1 bg-pg-danger text-white rounded text-xs hover:bg-pg-danger/90 disabled:opacity-50"
                                >
                                  {submitting ? '...' : 'Yes'}
                                </button>
                                <button
                                  onClick={() => setDeleteConfirm(null)}
                                  className="px-2 py-1 border border-pg-slate/20 text-pg-dark rounded text-xs hover:bg-pg-slate/5"
                                >
                                  No
                                </button>
                              </div>
                            ) : (
                              <button
                                onClick={() => setDeleteConfirm(String(user.id))}
                                className="px-3 py-1 text-pg-danger border border-pg-danger/20 rounded text-xs hover:bg-pg-danger/5 transition-colors flex items-center gap-1"
                              >
                                <Trash2 className="w-3 h-3" />
                                Delete
                              </button>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </>
            )}
          </div>

          {/* Role Descriptions */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-white rounded-lg shadow p-6">
              <h4 className="font-semibold text-pg-dark mb-2 flex items-center gap-2">
                <span className="w-3 h-3 rounded-full bg-pg-danger" />
                Admin
              </h4>
              <p className="text-sm text-pg-slate">
                Full access to all features. Can manage users, collectors, and system settings.
              </p>
            </div>
            <div className="bg-white rounded-lg shadow p-6">
              <h4 className="font-semibold text-pg-dark mb-2 flex items-center gap-2">
                <span className="w-3 h-3 rounded-full bg-pg-warning" />
                User
              </h4>
              <p className="text-sm text-pg-slate">
                Standard access. Can view dashboards, manage collectors, and acknowledge alerts.
              </p>
            </div>
            <div className="bg-white rounded-lg shadow p-6">
              <h4 className="font-semibold text-pg-dark mb-2 flex items-center gap-2">
                <span className="w-3 h-3 rounded-full bg-pg-blue" />
                Viewer
              </h4>
              <p className="text-sm text-pg-slate">
                Read-only access to dashboards, reports, and metrics. Cannot make changes.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* API Tokens Tab */}
      {activeTab === 'tokens' && (
        <div className="space-y-6">
          {/* New Token Form */}
          {showNewTokenForm && (
            <div className="bg-white rounded-lg shadow p-6 border-l-4 border-pg-blue">
              <h3 className="text-lg font-semibold text-pg-dark mb-4">Generate New API Token</h3>
              <form onSubmit={handleAddToken}>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-pg-dark mb-2">Token Name</label>
                  <input
                    type="text"
                    placeholder="e.g., Production Monitor"
                    value={newToken.name}
                    onChange={(e) => setNewToken({ ...newToken, name: e.target.value })}
                    className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                    required
                  />
                  <p className="text-xs text-pg-slate mt-2">
                    Use descriptive names for easy identification
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    type="submit"
                    className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors text-sm font-medium"
                  >
                    Generate Token
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowNewTokenForm(false)}
                    className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Tokens Table */}
          <div className="bg-white rounded-lg shadow p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-pg-dark">API Tokens</h3>
              <button
                onClick={() => setShowNewTokenForm(!showNewTokenForm)}
                className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
              >
                <Plus className="w-4 h-4" />
                Generate Token
              </button>
            </div>
            <DataTable
              columns={tokenColumns}
              data={mockTokens}
              searchable={true}
              emptyMessage="No API tokens found"
            />
          </div>

          {/* Token Display */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-pg-dark mb-4">Token Details</h3>
            <div className="space-y-3">
              {mockTokens.map((token) => (
                <div key={token.id} className="flex items-center justify-between p-4 bg-pg-slate/5 rounded-lg">
                  <div>
                    <h4 className="font-medium text-pg-dark">{token.name}</h4>
                    <p className="text-xs text-pg-slate">ID: {token.id}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <input
                      type={visibleTokens.has(token.id) ? 'text' : 'password'}
                      value={token.token}
                      readOnly
                      className="px-3 py-2 bg-white border border-pg-slate/20 rounded text-sm font-mono w-64"
                    />
                    <button
                      onClick={() => toggleTokenVisibility(token.id)}
                      className="p-2 hover:bg-pg-slate/10 rounded transition-colors"
                    >
                      {visibleTokens.has(token.id) ? (
                        <EyeOff className="w-4 h-4 text-pg-slate" />
                      ) : (
                        <Eye className="w-4 h-4 text-pg-slate" />
                      )}
                    </button>
                    <button
                      onClick={() => copyToClipboard(token.token, token.id)}
                      className="p-2 hover:bg-pg-slate/10 rounded transition-colors"
                    >
                      {copiedId === token.id ? (
                        <Check className="w-4 h-4 text-pg-success" />
                      ) : (
                        <Copy className="w-4 h-4 text-pg-slate" />
                      )}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Notifications Tab */}
      {activeTab === 'notifications' && (
        <div className="space-y-6">
          {mockNotifications.map((channel) => (
            <div key={channel.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-pg-dark flex items-center gap-2">
                    {channel.type === 'email' && <Bell className="w-5 h-5 text-pg-blue" />}
                    {channel.type === 'slack' && <Bell className="w-5 h-5 text-pg-blue" />}
                    {channel.type === 'pagerduty' && <AlertTriangle className="w-5 h-5 text-pg-warning" />}
                    {channel.type === 'webhook' && <SettingsIcon className="w-5 h-5 text-pg-slate" />}
                    {channel.name}
                  </h3>
                  <p className="text-xs text-pg-slate mt-1">{channel.type.toUpperCase()}</p>
                </div>
                <div className="flex items-center gap-3">
                  {channel.test_status && (
                    <StatusBadge
                      status={
                        channel.test_status === 'success'
                          ? 'success'
                          : channel.test_status === 'failed'
                            ? 'error'
                            : 'warning'
                      }
                      label={channel.test_status.toUpperCase()}
                      size="sm"
                    />
                  )}
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" checked={channel.enabled} readOnly className="rounded" />
                    <span className="text-sm text-pg-slate">
                      {channel.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </label>
                </div>
              </div>

              <div className="bg-pg-slate/5 rounded p-4 mb-4">
                {channel.type === 'email' && (
                  <div>
                    <p className="text-sm text-pg-slate">
                      <strong>Email:</strong> {channel.config.email}
                    </p>
                  </div>
                )}
                {channel.type === 'slack' && (
                  <div>
                    <p className="text-sm text-pg-slate">
                      <strong>Channel:</strong> {channel.config.slack_channel}
                    </p>
                  </div>
                )}
                {channel.type === 'webhook' && (
                  <div>
                    <p className="text-sm text-pg-slate font-mono text-xs break-all">
                      {channel.config.webhook_url}
                    </p>
                  </div>
                )}
              </div>

              <div className="flex gap-2">
                <button className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium">
                  Test Connection
                </button>
                <button className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium">
                  Edit
                </button>
                {deleteConfirm === channel.id ? (
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-pg-danger font-medium">Delete?</span>
                    <button
                      onClick={() => {
                        console.log('Deleting:', channel.id);
                        setDeleteConfirm(null);
                      }}
                      className="px-3 py-1 bg-pg-danger text-white rounded text-sm hover:bg-pg-danger/90 transition-colors"
                    >
                      Delete
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(null)}
                      className="px-3 py-1 border border-pg-slate/20 text-pg-dark rounded text-sm hover:bg-pg-slate/5 transition-colors"
                    >
                      Cancel
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setDeleteConfirm(channel.id)}
                    className="px-4 py-2 text-pg-danger border border-pg-danger/20 rounded-lg hover:bg-pg-danger/5 transition-colors flex items-center gap-2 text-sm font-medium"
                  >
                    <Trash2 className="w-4 h-4" />
                    Delete
                  </button>
                )}
              </div>
            </div>
          ))}

          {/* Add New Channel */}
          <div className="bg-white rounded-lg shadow p-6 border-l-4 border-pg-blue">
            <h3 className="text-lg font-semibold text-pg-dark mb-4">Add Notification Channel</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Channel Type</label>
                <select className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue">
                  <option>Select a type</option>
                  <option>Email</option>
                  <option>Slack</option>
                  <option>PagerDuty</option>
                  <option>Webhook</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Name</label>
                <input
                  type="text"
                  placeholder="e.g., Team Alerts"
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors text-sm font-medium">
                Add Channel
              </button>
            </div>
          </div>
        </div>
      )}
    </PageWrapper>
  );
};
