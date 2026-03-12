import { MainLayout } from '../components/layout/MainLayout'
import { ChannelsViewer } from '../components/channels/ChannelsViewer'

export const ChannelsPage: React.FC = () => {
  return (
    <MainLayout>
      <div className="space-y-6 p-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            Notification Channels
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            Configure notification channels for alert delivery
          </p>
        </div>

        <ChannelsViewer />
      </div>
    </MainLayout>
  )
}
