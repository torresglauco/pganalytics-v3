import React from 'react'
import { Link } from 'react-router-dom'

interface DrillDownCardProps {
  icon: string
  title: string
  description: string
  href: string
}

export const DrillDownCard: React.FC<DrillDownCardProps> = ({
  icon,
  title,
  description,
  href,
}) => {
  return (
    <Link
      to={href}
      className="block p-6 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-md hover:border-primary-300 dark:hover:border-primary-500 transition-all duration-200 group"
    >
      <div className="flex items-start gap-4">
        <span className="text-4xl group-hover:scale-110 transition-transform">
          {icon}
        </span>
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-1">
            {title}
          </h3>
          <p className="text-slate-600 dark:text-slate-400 text-sm">
            {description}
          </p>
          <span className="inline-block mt-3 text-primary-600 hover:text-primary-700 font-medium text-sm">
            Explore →
          </span>
        </div>
      </div>
    </Link>
  )
}

DrillDownCard.displayName = 'DrillDownCard'
