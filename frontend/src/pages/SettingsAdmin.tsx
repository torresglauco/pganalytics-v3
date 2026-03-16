import React, { useState, useMemo, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
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
import { MainLayout } from '../components/layout/MainLayout';
import { PageWrapper } from '../components/common/PageWrapper';
import { StatusBadge } from '../components/cards/StatusBadge';
import { DataTable, Column } from '../components/tables/DataTable';
import { formatDateTime } from '../utils/formatting';
import { useUsers, type AdminUser } from '../hooks/useUsers';
import { useApiTokens } from '../hooks/useApiTokens';
import { useChannels } from '../hooks/useChannels';

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

export const SettingsAdmin: React.FC = () => {
  const location = useLocation();
  const [activeTab, setActiveTab] = useState<'users' | 'tokens' | 'notifications'>('users');
  const [showNewUserForm, setShowNewUserForm] = useState(false);
  const [showNewTokenForm, setShowNewTokenForm] = useState(false);
  const [showNewChannelForm, setShowNewChannelForm] = useState(false);
  const [visibleTokens, setVisibleTokens] = useState<Set<string>>(new Set());
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [deleteChannelConfirm, setDeleteChannelConfirm] = useState<string | null>(null);
  const [resetPasswordConfirm, setResetPasswordConfirm] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [testingChannelId, setTestingChannelId] = useState<string | null>(null);

  // Set active tab based on current route
  useEffect(() => {
    if (location.pathname === '/users') {
      setActiveTab('users');
    } else if (location.pathname === '/settings') {
      setActiveTab('tokens');
    }
  }, [location.pathname]);

  // Use the useUsers hook
  const { users, loading, error, fetchUsers, createUser, deleteUser, resetPassword } = useUsers();

  // Use the useApiTokens hook
  const { data: tokens, loading: tokensLoading, error: tokensError, fetchTokens, createToken, deleteToken } = useApiTokens();

  // Use the useChannels hook
  const { data: channels, loading: channelsLoading, error: channelsError, fetchChannels, createChannel, deleteChannel, testChannel } = useChannels();

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

  const handleAddToken = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newToken.name) {
      setErrorMessage('Token name is required');
      return;
    }

    setSubmitting(true);
    try {
      await createToken({
        name: newToken.name,
      });
      setSuccessMessage('API token created successfully');
      setShowNewTokenForm(false);
      setNewToken({ name: '' });
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to create token');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteToken = async (tokenId: string) => {
    setSubmitting(true);
    try {
      await deleteToken(tokenId);
      setSuccessMessage('Token deleted successfully');
      setDeleteConfirm(null);
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to delete token');
    } finally {
      setSubmitting(false);
    }
  };

  interface NewChannelForm {
    type: 'email' | 'slack' | 'pagerduty' | 'webhook' | '';
    name: string;
    email?: string;
    slack_channel?: string;
    webhook_url?: string;
  }

  const [newChannel, setNewChannel] = useState<NewChannelForm>({
    type: '',
    name: '',
  });

  const handleAddChannel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newChannel.type || !newChannel.name) {
      setErrorMessage('Channel type and name are required');
      return;
    }

    const config: Record<string, string> = {};
    if (newChannel.type === 'email' && newChannel.email) {
      config.email = newChannel.email;
    } else if (newChannel.type === 'slack' && newChannel.slack_channel) {
      config.slack_channel = newChannel.slack_channel;
    } else if (newChannel.type === 'webhook' && newChannel.webhook_url) {
      config.webhook_url = newChannel.webhook_url;
    } else {
      setErrorMessage('Please provide the required configuration for this channel type');
      return;
    }

    setSubmitting(true);
    try {
      await createChannel({
        type: newChannel.type,
        name: newChannel.name,
        config,
        enabled: true,
      });
      setSuccessMessage('Notification channel created successfully');
      setShowNewChannelForm(false);
      setNewChannel({ type: '', name: '' });
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to create channel');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteChannel = async (channelId: string) => {
    setSubmitting(true);
    try {
      await deleteChannel(channelId);
      setSuccessMessage('Channel deleted successfully');
      setDeleteChannelConfirm(null);
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to delete channel');
    } finally {
      setSubmitting(false);
    }
  };

  const handleTestChannel = async (channelId: string) => {
    setTestingChannelId(channelId);
    try {
      await testChannel(channelId);
      setSuccessMessage('Channel test sent successfully');
    } catch (err) {
      const error = err as any;
      setErrorMessage(error.message || 'Failed to test channel');
    } finally {
      setTestingChannelId(null);
    }
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
    <MainLayout>
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
          {/* Quick Actions */}
          {users.length === 0 && !showNewUserForm && (
            <div className="bg-gradient-to-r from-purple-50 to-pink-50 border-2 border-purple-300 rounded-lg p-8">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-2xl font-bold text-purple-900">👥 Add Users</h3>
                  <p className="text-base text-purple-700 mt-2">Invite team members to manage PostgreSQL databases.</p>
                </div>
                <button
                  onClick={() => setShowNewUserForm(true)}
                  className="flex items-center gap-2 px-8 py-4 bg-purple-600 text-white rounded-lg hover:bg-purple-700 active:bg-purple-800 transition-all font-bold whitespace-nowrap shadow-lg text-lg"
                >
                  <Plus className="w-6 h-6" />
                  Add User
                </button>
              </div>
            </div>
          )}

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
          {/* Quick Actions */}
          {tokens.length === 0 && !showNewTokenForm && (
            <div className="bg-gradient-to-r from-amber-50 to-orange-50 border-2 border-amber-300 rounded-lg p-8">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-2xl font-bold text-amber-900">🔑 Generate API Token</h3>
                  <p className="text-base text-amber-700 mt-2">Create API tokens for programmatic access to pgAnalytics.</p>
                </div>
                <button
                  onClick={() => setShowNewTokenForm(true)}
                  className="flex items-center gap-2 px-8 py-4 bg-amber-600 text-white rounded-lg hover:bg-amber-700 active:bg-amber-800 transition-all font-bold whitespace-nowrap shadow-lg text-lg"
                >
                  <Plus className="w-6 h-6" />
                  Generate Token
                </button>
              </div>
            </div>
          )}

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
                    disabled={submitting}
                    className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 disabled:bg-pg-slate/50 transition-colors text-sm font-medium"
                  >
                    {submitting ? 'Generating...' : 'Generate Token'}
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
              <div className="flex gap-2">
                <button
                  onClick={() => fetchTokens()}
                  disabled={tokensLoading}
                  className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium disabled:opacity-50"
                  title="Refresh tokens list"
                >
                  <RotateCw className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setShowNewTokenForm(!showNewTokenForm)}
                  className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
                >
                  <Plus className="w-4 h-4" />
                  Generate Token
                </button>
              </div>
            </div>
            {tokensLoading ? (
              <div className="text-center py-8">
                <p className="text-pg-slate">Loading tokens...</p>
              </div>
            ) : (
              <>
                <DataTable
                  columns={tokenColumns}
                  data={tokens.map((token) => ({
                    ...token,
                    created_at: token.created_at ? new Date(token.created_at) : new Date(),
                    last_used: token.last_used ? new Date(token.last_used) : null,
                    expires_at: token.expires_at ? new Date(token.expires_at) : null,
                  }))}
                  searchable={true}
                  emptyMessage="No API tokens found"
                />
              </>
            )}
          </div>

          {/* Token Display and Management */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-pg-dark mb-4">Token Management</h3>
            {tokensLoading ? (
              <div className="text-center py-8">
                <p className="text-pg-slate">Loading tokens...</p>
              </div>
            ) : tokens.length === 0 ? (
              <p className="text-pg-slate text-center py-8">No API tokens created yet</p>
            ) : (
              <div className="space-y-3">
                {tokens.map((token) => (
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
                      {deleteConfirm === token.id ? (
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-pg-danger font-medium">Delete?</span>
                          <button
                            onClick={() => handleDeleteToken(token.id)}
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
                          onClick={() => setDeleteConfirm(token.id)}
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
        </div>
      )}

      {/* Notifications Tab */}
      {activeTab === 'notifications' && (
        <div className="space-y-6">
          {/* Channels List */}
          {channelsLoading ? (
            <div className="text-center py-8">
              <p className="text-pg-slate">Loading notification channels...</p>
            </div>
          ) : (channels || []).length === 0 ? (
            <div className="bg-white rounded-lg shadow p-6 text-center">
              <p className="text-pg-slate">No notification channels configured yet</p>
            </div>
          ) : (
            (channels || []).map((channel: any) => (
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
                        <strong>Email:</strong> {channel.config?.email}
                      </p>
                    </div>
                  )}
                  {channel.type === 'slack' && (
                    <div>
                      <p className="text-sm text-pg-slate">
                        <strong>Channel:</strong> {channel.config?.slack_channel}
                      </p>
                    </div>
                  )}
                  {channel.type === 'webhook' && (
                    <div>
                      <p className="text-sm text-pg-slate font-mono text-xs break-all">
                        {channel.config?.webhook_url}
                      </p>
                    </div>
                  )}
                </div>

                <div className="flex gap-2">
                  <button
                    onClick={() => handleTestChannel(channel.id)}
                    disabled={testingChannelId === channel.id}
                    className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium disabled:opacity-50"
                  >
                    {testingChannelId === channel.id ? 'Testing...' : 'Test Connection'}
                  </button>
                  {deleteChannelConfirm === channel.id ? (
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-pg-danger font-medium">Delete?</span>
                      <button
                        onClick={() => handleDeleteChannel(channel.id)}
                        disabled={submitting}
                        className="px-3 py-1 bg-pg-danger text-white rounded text-sm hover:bg-pg-danger/90 transition-colors disabled:opacity-50"
                      >
                        {submitting ? '...' : 'Delete'}
                      </button>
                      <button
                        onClick={() => setDeleteChannelConfirm(null)}
                        className="px-3 py-1 border border-pg-slate/20 text-pg-dark rounded text-sm hover:bg-pg-slate/5 transition-colors"
                      >
                        Cancel
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setDeleteChannelConfirm(channel.id)}
                      className="px-4 py-2 text-pg-danger border border-pg-danger/20 rounded-lg hover:bg-pg-danger/5 transition-colors flex items-center gap-2 text-sm font-medium"
                    >
                      <Trash2 className="w-4 h-4" />
                      Delete
                    </button>
                  )}
                </div>
              </div>
            ))
          )}

          {/* Add New Channel Form */}
          {showNewChannelForm && (
            <div className="bg-white rounded-lg shadow p-6 border-l-4 border-pg-blue">
              <h3 className="text-lg font-semibold text-pg-dark mb-4">Add Notification Channel</h3>
              <form onSubmit={handleAddChannel}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Channel Type</label>
                    <select
                      value={newChannel.type}
                      onChange={(e) => setNewChannel({ ...newChannel, type: e.target.value as any })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    >
                      <option value="">Select a type</option>
                      <option value="email">Email</option>
                      <option value="slack">Slack</option>
                      <option value="pagerduty">PagerDuty</option>
                      <option value="webhook">Webhook</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-pg-dark mb-2">Name</label>
                    <input
                      type="text"
                      placeholder="e.g., Team Alerts"
                      value={newChannel.name}
                      onChange={(e) => setNewChannel({ ...newChannel, name: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                </div>

                {/* Conditional fields based on channel type */}
                {newChannel.type === 'email' && (
                  <div className="mb-4">
                    <label className="block text-sm font-medium text-pg-dark mb-2">Email Address</label>
                    <input
                      type="email"
                      placeholder="alerts@company.com"
                      value={newChannel.email || ''}
                      onChange={(e) => setNewChannel({ ...newChannel, email: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                )}

                {newChannel.type === 'slack' && (
                  <div className="mb-4">
                    <label className="block text-sm font-medium text-pg-dark mb-2">Slack Channel</label>
                    <input
                      type="text"
                      placeholder="#alerts"
                      value={newChannel.slack_channel || ''}
                      onChange={(e) => setNewChannel({ ...newChannel, slack_channel: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                )}

                {newChannel.type === 'webhook' && (
                  <div className="mb-4">
                    <label className="block text-sm font-medium text-pg-dark mb-2">Webhook URL</label>
                    <input
                      type="url"
                      placeholder="https://api.example.com/webhooks"
                      value={newChannel.webhook_url || ''}
                      onChange={(e) => setNewChannel({ ...newChannel, webhook_url: e.target.value })}
                      className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                      required
                    />
                  </div>
                )}

                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 disabled:bg-pg-slate/50 transition-colors text-sm font-medium"
                  >
                    {submitting ? 'Adding...' : 'Add Channel'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowNewChannelForm(false)}
                    className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Add New Channel Button */}
          {!showNewChannelForm && (
            <div className="bg-white rounded-lg shadow p-6">
              <button
                onClick={() => setShowNewChannelForm(true)}
                className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
              >
                <Plus className="w-4 h-4" />
                Add Notification Channel
              </button>
            </div>
          )}
        </div>
      )}
    </PageWrapper>
    </MainLayout>
  );
};
