import { MainLayout } from '../components/layout/MainLayout'
import { MetricsViewer } from '../components/metrics/MetricsViewer'

export const MetricsPage: React.FC = () => {
  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            Metrics & Analytics
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            View PostgreSQL performance metrics and error trends
          </p>
        </div>

        <MetricsViewer />
      </div>
    </MainLayout>
  )
}
