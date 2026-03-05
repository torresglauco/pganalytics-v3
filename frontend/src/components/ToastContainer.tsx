import React from 'react';
import {
  CheckCircle,
  AlertCircle,
  AlertTriangle,
  Info,
  X,
} from 'lucide-react';
import { useToast } from '../contexts/ToastContext';

export const ToastContainer: React.FC = () => {
  const { toasts, removeToast } = useToast();

  const getIcon = (type: string) => {
    switch (type) {
      case 'success':
        return <CheckCircle size={20} className="text-green-600" />;
      case 'error':
        return <AlertCircle size={20} className="text-red-600" />;
      case 'warning':
        return <AlertTriangle size={20} className="text-yellow-600" />;
      case 'info':
        return <Info size={20} className="text-blue-600" />;
      default:
        return null;
    }
  };

  const getStyles = (type: string) => {
    switch (type) {
      case 'success':
        return 'bg-green-50 border-green-200 text-green-800';
      case 'error':
        return 'bg-red-50 border-red-200 text-red-800';
      case 'warning':
        return 'bg-yellow-50 border-yellow-200 text-yellow-800';
      case 'info':
        return 'bg-blue-50 border-blue-200 text-blue-800';
      default:
        return 'bg-gray-50 border-gray-200 text-gray-800';
    }
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-sm">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`border rounded-lg p-4 shadow-lg animate-in fade-in slide-in-from-top-2 ${getStyles(
            toast.type
          )}`}
        >
          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 pt-0.5">{getIcon(toast.type)}</div>

            <div className="flex-1 min-w-0">
              {toast.title && (
                <h3 className="font-semibold mb-1">{toast.title}</h3>
              )}
              <p className="text-sm">{toast.message}</p>

              {toast.action && (
                <button
                  onClick={() => {
                    toast.action?.onClick();
                    removeToast(toast.id);
                  }}
                  className="text-sm font-medium mt-2 underline hover:opacity-80"
                >
                  {toast.action.label}
                </button>
              )}
            </div>

            <button
              onClick={() => removeToast(toast.id)}
              className="flex-shrink-0 opacity-60 hover:opacity-100 transition"
            >
              <X size={18} />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
};

export default ToastContainer;
