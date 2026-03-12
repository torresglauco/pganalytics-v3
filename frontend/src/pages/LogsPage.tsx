import { MainLayout } from '../components/layout/MainLayout'
import { LogsViewer } from '../components/logs/LogsViewer'

export const LogsPage: React.FC = () => {
  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            PostgreSQL Logs
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            View and filter PostgreSQL logs across all instances
          </p>
        </div>

        <LogsViewer />
      </div>
    </MainLayout>
  )
}
