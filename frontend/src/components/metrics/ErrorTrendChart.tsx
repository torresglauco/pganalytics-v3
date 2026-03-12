import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'

interface ErrorTrendChartProps {
  data: any[]
  loading: boolean
}

export const ErrorTrendChart: React.FC<ErrorTrendChartProps> = ({ data, loading }) => {
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
        Error Trend (24h)
      </h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
          <XAxis
            dataKey="timestamp"
            stroke="#94a3b8"
            style={{ fontSize: '12px' }}
          />
          <YAxis stroke="#94a3b8" style={{ fontSize: '12px' }} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #475569',
              borderRadius: '8px',
            }}
            labelStyle={{ color: '#e2e8f0' }}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="errors"
            stroke="#dc2626"
            dot={false}
            strokeWidth={2}
            name="Errors"
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
