import { useNavigate } from 'react-router-dom'

interface NotImplementedPageProps {
  title: string
  icon: string
  description: string
}

export const NotImplementedPage: React.FC<NotImplementedPageProps> = ({
  title,
  icon,
  description,
}) => {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
        {/* Icon */}
        <div className="text-6xl mb-4">{icon}</div>

        {/* Title */}
        <h1 className="text-3xl font-bold text-slate-900 mb-2">{title}</h1>

        {/* Description */}
        <p className="text-slate-600 mb-8">{description}</p>

        {/* Status Badge */}
        <div className="inline-block bg-yellow-50 border border-yellow-200 rounded-lg px-4 py-2 mb-8">
          <p className="text-sm font-medium text-yellow-800">
            ⏳ Coming in Phase 4.1
          </p>
        </div>

        {/* Info Box */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-8 text-left">
          <p className="text-sm text-blue-900">
            <span className="font-semibold">ℹ️ Info:</span> This page is currently
            being developed. Check back soon!
          </p>
        </div>

        {/* Back Button */}
        <button
          onClick={() => navigate('/')}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition"
        >
          ← Back to Dashboard
        </button>
      </div>
    </div>
  )
}
