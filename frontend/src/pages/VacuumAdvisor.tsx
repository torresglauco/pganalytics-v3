import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useVacuumAdvisor } from '../hooks/useVacuumAdvisor';
import { MainLayout } from '../components/layout/MainLayout';
import {
  VacuumRecommendation,
  AutovacuumConfig,
  AutovacuumTuningSuggestion,
  VacuumFilter,
  VacuumSort,
} from '../types/vacuumAdvisor';

export const VacuumAdvisorPage: React.FC = () => {
  const { databaseId } = useParams<{ databaseId: string }>();
  const numericDatabaseId = databaseId ? parseInt(databaseId, 10) : 0;
  const {
    recommendations,
    recommendationsLoading,
    recommendationsError,
    autovacuumConfigs,
    autovacuumConfigLoading,
    tuningSuggestions,
    tuningSuggestionsLoading,
    executeVacuum,
    vacuumExecuting,
    filter,
    setFilter,
    sort,
    setSort,
    filteredAndSortedRecommendations,
  } = useVacuumAdvisor(numericDatabaseId);

  const [selectedTab, setSelectedTab] = useState<'recommendations' | 'config' | 'tuning'>(
    'recommendations'
  );

  // Calculate summary metrics
  const summary = {
    total_tables: recommendations.length,
    tables_needing_vacuum: recommendations.filter(
      (r) => r.recommendation_type === 'full_vacuum'
    ).length,
    total_dead_space_bytes: recommendations.reduce((sum, r) => sum + r.estimated_gain, 0),
    average_dead_ratio:
      recommendations.length > 0
        ? recommendations.reduce((sum, r) => sum + r.dead_tuples_ratio, 0) / recommendations.length
        : 0,
    autovacuum_disabled_count: recommendations.filter((r) => !r.autovacuum_enabled).length,
  };

  const formatBytes = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
  };

  const formatDate = (date?: string): string => {
    if (!date) return 'Never';
    return new Date(date).toLocaleDateString();
  };

  const getRecommendationColor = (type: string): string => {
    switch (type) {
      case 'full_vacuum':
        return 'text-red-600';
      case 'tune_autovacuum':
        return 'text-yellow-600';
      case 'analyze_only':
        return 'text-blue-600';
      default:
        return 'text-gray-600';
    }
  };

  const getImpactBadgeColor = (impact: string): string => {
    switch (impact) {
      case 'high':
        return 'bg-red-100 text-red-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // IMPORTANT: This return statement has MainLayout wrapper
  return (
    <MainLayout>
      <div className="p-6 max-w-7xl mx-auto" data-testid="vacuum-advisor-inner">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">VACUUM Advisor</h1>
          <p className="text-gray-600 mt-2">
            Database maintenance recommendations for table bloat prevention
          </p>
        </div>

      {/* Summary Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-8">
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-sm font-medium text-gray-500">Total Tables</h3>
          <p className="text-2xl font-bold text-gray-900 mt-2">{summary.total_tables}</p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-sm font-medium text-gray-500">Needing Vacuum</h3>
          <p className="text-2xl font-bold text-red-600 mt-2">{summary.tables_needing_vacuum}</p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-sm font-medium text-gray-500">Space to Recover</h3>
          <p className="text-2xl font-bold text-gray-900 mt-2">
            {formatBytes(summary.total_dead_space_bytes)}
          </p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-sm font-medium text-gray-500">Avg Dead Ratio</h3>
          <p className="text-2xl font-bold text-gray-900 mt-2">{summary.average_dead_ratio.toFixed(1)}%</p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-sm font-medium text-gray-500">Autovacuum Disabled</h3>
          <p className="text-2xl font-bold text-yellow-600 mt-2">{summary.autovacuum_disabled_count}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="flex border-b">
          <button
            onClick={() => setSelectedTab('recommendations')}
            className={`flex-1 py-4 px-6 text-center font-medium ${
              selectedTab === 'recommendations'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            VACUUM Recommendations
          </button>
          <button
            onClick={() => setSelectedTab('config')}
            className={`flex-1 py-4 px-6 text-center font-medium ${
              selectedTab === 'config'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Autovacuum Config
          </button>
          <button
            onClick={() => setSelectedTab('tuning')}
            className={`flex-1 py-4 px-6 text-center font-medium ${
              selectedTab === 'tuning'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Tuning Suggestions
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* VACUUM Recommendations Tab */}
          {selectedTab === 'recommendations' && (
            <div>
              {/* Filters */}
              <div className="mb-6 flex gap-4 flex-wrap">
                <select
                  value={filter.recommendation_type || ''}
                  onChange={(e) =>
                    setFilter({
                      ...filter,
                      recommendation_type: (e.target.value as any) || undefined,
                    })
                  }
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="">All Types</option>
                  <option value="full_vacuum">Full VACUUM</option>
                  <option value="analyze_only">Analyze Only</option>
                  <option value="tune_autovacuum">Tune Autovacuum</option>
                </select>

                <select
                  value={sort.field}
                  onChange={(e) =>
                    setSort({ ...sort, field: e.target.value as any })
                  }
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="dead_ratio">Sort by Dead Ratio</option>
                  <option value="estimated_gain">Sort by Gain</option>
                  <option value="table_size">Sort by Table Size</option>
                  <option value="last_vacuum">Sort by Last Vacuum</option>
                </select>

                <button
                  onClick={() =>
                    setSort({ ...sort, order: sort.order === 'asc' ? 'desc' : 'asc' })
                  }
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm hover:bg-gray-50"
                >
                  {sort.order === 'asc' ? '↑ Ascending' : '↓ Descending'}
                </button>
              </div>

              {/* Loading state */}
              {recommendationsLoading && <p className="text-gray-600">Loading recommendations...</p>}

              {/* Error state */}
              {recommendationsError && (
                <p className="text-red-600">Error: {recommendationsError}</p>
              )}

              {/* Recommendations table */}
              {!recommendationsLoading && filteredAndSortedRecommendations.length > 0 && (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="border-b">
                      <tr>
                        <th className="text-left py-3 px-4">Table Name</th>
                        <th className="text-left py-3 px-4">Size</th>
                        <th className="text-left py-3 px-4">Dead Ratio</th>
                        <th className="text-left py-3 px-4">Estimated Gain</th>
                        <th className="text-left py-3 px-4">Recommendation</th>
                        <th className="text-left py-3 px-4">Last Vacuum</th>
                        <th className="text-left py-3 px-4">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredAndSortedRecommendations.map((rec) => (
                        <tr key={rec.id} className="border-b hover:bg-gray-50">
                          <td className="py-3 px-4 font-medium">{rec.table_name}</td>
                          <td className="py-3 px-4">{formatBytes(rec.table_size)}</td>
                          <td className="py-3 px-4">{rec.dead_tuples_ratio.toFixed(2)}%</td>
                          <td className="py-3 px-4">{formatBytes(rec.estimated_gain)}</td>
                          <td className={`py-3 px-4 font-medium ${getRecommendationColor(rec.recommendation_type)}`}>
                            {rec.recommendation_type.replace(/_/g, ' ')}
                          </td>
                          <td className="py-3 px-4">{formatDate(rec.last_vacuum)}</td>
                          <td className="py-3 px-4">
                            {rec.recommendation_type === 'full_vacuum' && (
                              <button
                                onClick={() => executeVacuum(rec.id)}
                                disabled={vacuumExecuting}
                                className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
                              >
                                {vacuumExecuting ? 'Running...' : 'Execute'}
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {!recommendationsLoading && filteredAndSortedRecommendations.length === 0 && (
                <p className="text-gray-600">No recommendations found for the selected filters.</p>
              )}
            </div>
          )}

          {/* Autovacuum Config Tab */}
          {selectedTab === 'config' && (
            <div>
              {autovacuumConfigLoading && <p className="text-gray-600">Loading configurations...</p>}

              {autovacuumConfigs.length > 0 && (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="border-b">
                      <tr>
                        <th className="text-left py-3 px-4">Table Name</th>
                        <th className="text-left py-3 px-4">Setting</th>
                        <th className="text-left py-3 px-4">Current</th>
                        <th className="text-left py-3 px-4">Recommended</th>
                        <th className="text-left py-3 px-4">Impact</th>
                      </tr>
                    </thead>
                    <tbody>
                      {autovacuumConfigs.map((config) => (
                        <tr key={config.id} className="border-b hover:bg-gray-50">
                          <td className="py-3 px-4">{config.table_name}</td>
                          <td className="py-3 px-4">{config.setting_name}</td>
                          <td className="py-3 px-4 font-mono text-xs">{config.current_value}</td>
                          <td className="py-3 px-4 font-mono text-xs">{config.recommended_value}</td>
                          <td className="py-3 px-4">
                            <span className={`px-2 py-1 rounded text-xs font-medium ${getImpactBadgeColor(config.impact)}`}>
                              {config.impact}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {!autovacuumConfigLoading && autovacuumConfigs.length === 0 && (
                <p className="text-gray-600">No autovacuum configurations found.</p>
              )}
            </div>
          )}

          {/* Tuning Suggestions Tab */}
          {selectedTab === 'tuning' && (
            <div>
              {tuningSuggestionsLoading && <p className="text-gray-600">Loading tuning suggestions...</p>}

              {tuningSuggestions.length > 0 && (
                <div className="space-y-4">
                  {tuningSuggestions.map((suggestion, idx) => (
                    <div key={idx} className="border rounded-lg p-4 bg-gray-50">
                      <h3 className="font-medium text-gray-900 mb-2">
                        {suggestion.parameter.replace(/_/g, ' ')} for {suggestion.table_name}
                      </h3>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-3">
                        <div>
                          <p className="text-xs text-gray-600">Current</p>
                          <p className="font-mono text-sm">{suggestion.current_value}</p>
                        </div>
                        <div>
                          <p className="text-xs text-gray-600">Recommended</p>
                          <p className="font-mono text-sm">{suggestion.recommended_value}</p>
                        </div>
                      </div>
                      <p className="text-sm text-gray-700 mb-2">{suggestion.rationale}</p>
                      <p className="text-xs text-green-700 font-medium">
                        Expected Improvement: {suggestion.expected_improvement}%
                      </p>
                    </div>
                  ))}
                </div>
              )}

              {!tuningSuggestionsLoading && tuningSuggestions.length === 0 && (
                <p className="text-gray-600">No tuning suggestions available.</p>
              )}
            </div>
          )}
        </div>
      </div>
      </div>
    </MainLayout>
  );
};

export default VacuumAdvisorPage;
