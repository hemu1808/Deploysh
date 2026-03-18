
import React, { useState } from 'react';
import Header from './components/Header';
import Dashboard from './components/Dashboard';
import DeployAppModal from './components/DeployAppModal';
import { useApplications } from './hooks/useApplications';
import { ToastProvider, useToast } from './components/Toast';
import type { Application } from './types';
import AppDetailsModal from './components/AppDetailsModal';

const AppContent: React.FC = () => {
  const {
    applications,
    deployApplication,
    removeApplication,
    scaleApplication,
    isLoading
  } = useApplications();
  const { showToast } = useToast();
  const [isDeployModalOpen, setIsDeployModalOpen] = useState(false);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);

  const [searchQuery, setSearchQuery] = useState('');

  const handleDeploy = (image: string, replicas: number, envVars?: any[], ports?: any[]) => {
    deployApplication({ image, replicas, envVars, ports });
    setIsDeployModalOpen(false);
    showToast(`Successfully initiated deployment for ${image}`, 'success');
  };
  
  const handleOpenDetails = (app: Application) => {
    setSelectedApp(app);
  };
  
  const handleCloseDetails = () => {
    setSelectedApp(null);
  };

  return (
    <div className="min-h-screen font-sans">
      <Header onDeployClick={() => setIsDeployModalOpen(true)} />
      <main className="p-4 sm:p-6 lg:p-8">
        {isLoading ? (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
            </div>
        ) : (
            <div className="space-y-6">
                <div className="flex justify-between items-center mb-6">
                    <h2 className="text-2xl font-bold text-slate-100 tracking-tight">Active Deployments</h2>
                    <div className="relative">
                        <input 
                            type="text" 
                            placeholder="Search applications..." 
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="bg-slate-800/50 border border-glass-border text-slate-100 text-sm rounded-lg focus:ring-brand-blue focus:border-brand-blue block w-64 p-2.5 backdrop-blur-sm"
                        />
                    </div>
                </div>
                <Dashboard 
                  applications={(applications ?? []).filter(app => app.name.toLowerCase().includes(searchQuery.toLowerCase()) || app.dockerImage.toLowerCase().includes(searchQuery.toLowerCase()))}
                  onScale={(id, rep) => {
                      scaleApplication({appId: id, replicas: rep});
                      showToast(`Scaling application to ${rep} replicas`, 'info');
                  }}
                  onRemove={(id) => {
                      removeApplication(id);
                      showToast('Application removed', 'error');
                  }}
                  onViewDetails={handleOpenDetails}
                />
            </div>
        )}
      </main>
      {isDeployModalOpen && (
        <DeployAppModal
          onClose={() => setIsDeployModalOpen(false)}
          onDeploy={handleDeploy}
        />
      )}
      {selectedApp && (
        <AppDetailsModal
          app={selectedApp}
          onClose={handleCloseDetails}
        />
      )}
    </div>
    );
};

const App: React.FC = () => {
    return (
        <ToastProvider>
            <AppContent />
        </ToastProvider>
    );
};

export default App;
