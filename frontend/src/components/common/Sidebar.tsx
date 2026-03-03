import React from 'react';
import { ChevronRight, X } from 'lucide-react';
import { useUIStore } from '../../store/uiStore';
import { SIDEBAR_ITEMS } from '../../utils/constants';

interface SidebarProps {
  onNavigate: (pageId: string) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ onNavigate }) => {
  const sidebarOpen = useUIStore((state) => state.sidebarOpen);
  const currentPage = useUIStore((state) => state.currentPage);
  const setSidebarOpen = useUIStore((state) => state.setSidebarOpen);
  const setCurrentPage = useUIStore((state) => state.setCurrentPage);

  const handleNavigate = (pageId: string) => {
    setCurrentPage(pageId);
    onNavigate(pageId);
    // Close sidebar on mobile
    if (window.innerWidth < 768) {
      setSidebarOpen(false);
    }
  };

  return (
    <>
      {/* Overlay for mobile */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-30 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          fixed md:static inset-y-0 left-0 z-40
          w-64 bg-white border-r border-pg-slate/10
          transition-transform duration-300 ease-in-out
          ${sidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
          flex flex-col
        `}
      >
        {/* Close button for mobile */}
        <div className="md:hidden p-4 border-b border-pg-slate/10 flex justify-end">
          <button onClick={() => setSidebarOpen(false)}>
            <X className="w-5 h-5 text-pg-dark" />
          </button>
        </div>

        {/* Navigation items */}
        <nav className="flex-1 overflow-y-auto py-4">
          <div className="px-2">
            {SIDEBAR_ITEMS.map((item) => (
              <button
                key={item.id}
                onClick={() => handleNavigate(item.id)}
                className={`
                  w-full flex items-center gap-3 px-4 py-3 rounded-lg
                  transition-all duration-200
                  ${
                    currentPage === item.id
                      ? 'bg-pg-blue text-white shadow-md'
                      : 'text-pg-dark hover:bg-pg-slate/5'
                  }
                `}
              >
                <span className="text-lg">{item.icon}</span>
                <span className="font-medium flex-1 text-left">{item.label}</span>
                {currentPage === item.id && (
                  <ChevronRight className="w-4 h-4" />
                )}
              </button>
            ))}
          </div>
        </nav>

        {/* Footer info */}
        <div className="border-t border-pg-slate/10 p-4 bg-pg-slate/5">
          <p className="text-xs text-pg-slate text-center">
            pgAnalytics v3.3.0
          </p>
        </div>
      </aside>
    </>
  );
};
