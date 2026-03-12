import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
} from 'recharts'

interface LogDistributionChartProps {
  data: any[]
  loading: boolean
}

const COLORS = {
  DEBUG: '#64748b',
  INFO: '#3b82f6',
  NOTICE: '#0ea5e9',
  WARNING: '#f59e0b',
  ERROR: '#ef4444',
  FATAL: '#991b1b',
  PANIC: '#7c2d12',
}

export const LogDistributionChart: React.FC<LogDistributionChartProps> = ({
  data,
  loading,
}) => {
  if (loading) {
    return (
      <div className="h-80 w-full bg-slate-100 dark:bg-slate-800 rounded-lg flex items-center justify-center">
        <div className="text-slate-500">Loading chart...</div>
      </div>
    )
  }

  if (!data || data.length === 0) {
    return (
      <div className="h-80 w-full bg-slate-100 dark:bg-slate-800 rounded-lg flex items-center justify-center">
        <div className="text-slate-500">No data available</div>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-6">
      <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-4">
        Log Level Distribution
      </h3>
      <ResponsiveContainer width="100%" height={300}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
            outerRadius={100}
            fill="#8884d8"
            dataKey="value"
          >
            {data.map((entry: any) => (
              <Cell
                key={`cell-${entry.name}`}
                fill={COLORS[entry.name as keyof typeof COLORS] || '#64748b'}
              />
            ))}
          </Pie>
          <Tooltip />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}
