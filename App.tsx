
import React, { useState } from 'react';
import Header from './components/Header';
import Dashboard from './components/Dashboard';
import DeployAppModal from './components/DeployAppModal';
import { useMockData } from './hooks/useMockData';
import type { Application } from './types';
import AppDetailsModal from './components/AppDetailsModal';

const App: React.FC = () => {
  const {
    applications,
    addApplication,
    removeApplication,
    scaleApplication,
  } = useMockData();
  const [isDeployModalOpen, setIsDeployModalOpen] = useState(false);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);

  const handleDeploy = (image: string, replicas: number) => {
    addApplication(image, replicas);
    setIsDeployModalOpen(false);
  };
  
  const handleOpenDetails = (app: Application) => {
    setSelectedApp(app);
  };
  
  const handleCloseDetails = () => {
    setSelectedApp(null);
  };

  return (
    <div className="min-h-screen bg-gray-900 font-sans">
      <Header onDeployClick={() => setIsDeployModalOpen(true)} />
      <main className="p-4 sm:p-6 lg:p-8">
        <Dashboard 
          applications={applications}
          onScale={scaleApplication}
          onRemove={removeApplication}
          onViewDetails={handleOpenDetails}
        />
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

export default App;
