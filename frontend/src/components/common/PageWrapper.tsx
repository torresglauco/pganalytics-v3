import React from 'react';
import { ChevronRight } from 'lucide-react';

interface PageWrapperProps {
  title: string;
  description?: string;
  breadcrumbs?: Array<{ label: string; href?: string }>;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export const PageWrapper: React.FC<PageWrapperProps> = ({
  title,
  description,
  breadcrumbs,
  actions,
  children,
}) => {
  return (
    <div className="flex-1 overflow-auto bg-pg-light">
      {/* Page Header */}
      <div className="bg-white border-b border-pg-slate/10 sticky top-16 z-30">
        {/* Breadcrumbs */}
        {breadcrumbs && breadcrumbs.length > 0 && (
          <div className="px-6 py-2 flex items-center gap-2 text-sm">
            {breadcrumbs.map((crumb, index) => (
              <React.Fragment key={index}>
                {index > 0 && <ChevronRight className="w-4 h-4 text-pg-slate" />}
                <span className={crumb.href ? 'text-pg-cyan cursor-pointer hover:underline' : 'text-pg-slate'}>
                  {crumb.label}
                </span>
              </React.Fragment>
            ))}
          </div>
        )}

        {/* Title and Actions */}
        <div className="px-6 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-pg-dark">{title}</h1>
            {description && (
              <p className="text-pg-slate mt-1">{description}</p>
            )}
          </div>
          {actions && <div className="flex gap-2">{actions}</div>}
        </div>
      </div>

      {/* Page Content */}
      <main className="p-6">
        <div className="max-w-7xl mx-auto">
          {children}
        </div>
      </main>
    </div>
  );
};
