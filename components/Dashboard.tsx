
import React from 'react';
import type { Application } from '../types';
import AppCard from './AppCard';

interface DashboardProps {
  applications: Application[];
  onScale: (appId: string, newTargetReplicas: number) => void;
  onRemove: (appId: string) => void;
  onViewDetails: (app: Application) => void;
}

const Dashboard: React.FC<DashboardProps> = ({ applications, onScale, onRemove, onViewDetails }) => {
  if (applications.length === 0) {
    return (
      <div className="text-center py-20">
        <h2 className="text-2xl font-semibold text-gray-400">No Applications Deployed</h2>
        <p className="mt-2 text-gray-500">Click "Deploy New App" to get started.</p>
      </div>
    );
  }
  
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {applications.map(app => (
        <AppCard
          key={app.id}
          app={app}
          onScale={onScale}
          onRemove={onRemove}
          onViewDetails={onViewDetails}
        />
      ))}
    </div>
  );
};

export default Dashboard;
