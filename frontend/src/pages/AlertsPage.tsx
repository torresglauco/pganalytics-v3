import { MainLayout } from '../components/layout/MainLayout'
import { AlertsViewer } from '../components/alerts/AlertsViewer'

export const AlertsPage: React.FC = () => {
  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            Alert Rules
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            Create and manage alert rules for PostgreSQL monitoring
          </p>
        </div>

        <AlertsViewer />
      </div>
    </MainLayout>
  )
}
