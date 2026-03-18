import React, { useState } from 'react';
import type { Application } from '../types';
import StatusIndicator from './icons/StatusIndicator';
import MetricChart from './MetricChart';
import LogViewer from './LogViewer';
import { CloseIcon } from './icons/Icon';

interface AppDetailsModalProps {
  app: Application;
  onClose: () => void;
}

type Tab = 'overview' | 'metrics' | 'logs';

const AppDetailsModal: React.FC<AppDetailsModalProps> = ({ app, onClose }) => {
  const [activeTab, setActiveTab] = useState<Tab>('overview');

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-40" onClick={onClose}>
      <div 
        className="bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl h-[90vh] mx-4 flex flex-col" 
        onClick={e => e.stopPropagation()}
      >
        <header className="flex items-center justify-between p-4 border-b border-gray-700 flex-shrink-0">
          <div>
            <h2 className="text-xl font-bold text-gray-200">{app.name}</h2>
            <p className="text-sm text-gray-500">{app.dockerImage}</p>
          </div>
          <div className="flex items-center space-x-4">
            <StatusIndicator status={app.status} showText={true} />
            <button onClick={onClose} className="p-1 rounded-full text-gray-500 hover:bg-gray-700 hover:text-gray-300">
                <CloseIcon className="w-6 h-6" />
            </button>
          </div>
        </header>

        <div className="px-4 pt-4 border-b border-gray-700 flex-shrink-0">
          <nav className="flex space-x-4">
            <TabButton name="Overview" tab="overview" activeTab={activeTab} onClick={setActiveTab} />
            <TabButton name="Metrics" tab="metrics" activeTab={activeTab} onClick={setActiveTab} />
            <TabButton name="Logs" tab="logs" activeTab={activeTab} onClick={setActiveTab} />
          </nav>
        </div>

        <main className="flex-grow p-4 overflow-y-auto">
          {activeTab === 'overview' && (
              <div className="space-y-6">
                 <div>
                    <h3 className="text-sm font-semibold text-gray-400 mb-2 uppercase tracking-wide">Environment Variables</h3>
                    {app.envVars && app.envVars.length > 0 ? (
                        <div className="bg-gray-900 rounded-md border border-gray-700 overflow-hidden">
                            {app.envVars.map((env, i) => (
                                <div key={i} className="flex border-b border-gray-700 last:border-0 text-sm">
                                    <div className="w-1/3 p-2 bg-gray-800/50 font-mono text-gray-300 border-r border-gray-700">{env.key}</div>
                                    <div className="w-2/3 p-2 font-mono text-gray-400 truncate">{env.value}</div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-sm text-gray-500 italic">No environment variables injected.</p>
                    )}
                 </div>

                 <div>
                    <h3 className="text-sm font-semibold text-gray-400 mb-2 uppercase tracking-wide">Port Mappings</h3>
                    {app.ports && app.ports.length > 0 ? (
                        <div className="flex flex-wrap gap-2">
                             {app.ports.map((port, i) => (
                                 <span key={i} className="inline-flex items-center px-2.5 py-0.5 rounded-md text-sm font-medium bg-blue-900/30 text-blue-300 border border-blue-800">
                                     Host {port.hostPort} → Container {port.containerPort}
                                 </span>
                             ))}
                        </div>
                    ) : (
                        <p className="text-sm text-gray-500 italic">No port mappings configured.</p>
                    )}
                 </div>
              </div>
          )}
          {activeTab === 'metrics' && (
            <div className="space-y-6">
                <MetricChart title="CPU Usage (%)" data={app.cpuUsage} dataKey="value" color="#4299e1" />
                <MetricChart title="Memory Usage (MB)" data={app.memoryUsage} dataKey="value" color="#48bb78" />
            </div>
          )}
          {activeTab === 'logs' && <LogViewer logs={app.logs} />}
        </main>
      </div>
    </div>
  );
};

interface TabButtonProps {
    name: string;
    tab: Tab;
    activeTab: Tab;
    onClick: (tab: Tab) => void;
}

const TabButton: React.FC<TabButtonProps> = ({name, tab, activeTab, onClick}) => {
    const isActive = tab === activeTab;
    return (
        <button
            onClick={() => onClick(tab as any)}
            className={`px-3 py-2 font-medium text-sm rounded-t-md focus:outline-none transition-colors ${
            isActive
                ? 'bg-gray-800 border-gray-700 border-l border-t border-r text-gray-200'
                : 'text-gray-500 hover:text-gray-300'
            }`}
        >
            {name}
        </button>
    )
}

export default AppDetailsModal;
