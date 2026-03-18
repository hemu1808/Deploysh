import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { CheckCircleIcon, XCircleIcon, InformationalIcon, XIcon } from './icons/Icon';

type ToastType = 'success' | 'error' | 'info';

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

interface ToastContextType {
  showToast: (message: string, type: ToastType) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const ToastProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const showToast = useCallback((message: string, type: ToastType) => {
    const id = Date.now().toString();
    setToasts(prev => [...prev, { id, message, type }]);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id));
    }, 5000);
  }, []);

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col space-y-3 pointer-events-none">
        {toasts.map(toast => (
          <div 
            key={toast.id} 
            className={`glass-panel pointer-events-auto overflow-hidden rounded-lg shadow-lg max-w-sm w-full animate-[float_0.3s_ease-out] flex items-center p-4 border-l-4 ${
                toast.type === 'success' ? 'border-l-green-400 bg-green-500/10' :
                toast.type === 'error' ? 'border-l-red-400 bg-red-500/10' :
                'border-l-brand-blue bg-brand-blue/10'
            }`}
          >
            <div className="flex-shrink-0 mr-3">
              {toast.type === 'success' && <CheckCircleIcon className="w-5 h-5 text-green-400" />}
              {toast.type === 'error' && <XCircleIcon className="w-5 h-5 text-red-400" />}
              {toast.type === 'info' && <InformationalIcon className="w-5 h-5 text-brand-blue" />}
            </div>
            <div className="flex-1 w-0">
               <p className="text-sm font-medium text-slate-200">{toast.message}</p>
            </div>
            <div className="ml-4 flex-shrink-0 flex">
                <button
                  onClick={() => removeToast(toast.id)}
                  className="rounded-md inline-flex text-slate-400 hover:text-slate-200 focus:outline-none focus:ring-2 focus:ring-brand-blue"
                >
                  <span className="sr-only">Close</span>
                  <XIcon className="h-5 w-5" />
                </button>
            </div>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};
