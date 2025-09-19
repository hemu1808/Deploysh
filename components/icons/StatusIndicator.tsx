
import React from 'react';
import { ApplicationStatus } from '../../types';

interface StatusIndicatorProps {
  status: ApplicationStatus;
  showText?: boolean;
}

const StatusIndicator: React.FC<StatusIndicatorProps> = ({ status, showText = false }) => {
  const statusConfig = {
    [ApplicationStatus.HEALTHY]: { color: 'bg-green-500', text: 'Healthy' },
    [ApplicationStatus.UNHEALTHY]: { color: 'bg-red-500', text: 'Unhealthy' },
    [ApplicationStatus.DEPLOYING]: { color: 'bg-yellow-500', text: 'Deploying' },
    [ApplicationStatus.STOPPED]: { color: 'bg-gray-600', text: 'Stopped' },
  };

  const { color, text } = statusConfig[status];
  const isDeploying = status === ApplicationStatus.DEPLOYING;

  return (
    <div className="flex items-center space-x-2">
      <span className={`h-3 w-3 rounded-full ${color} ${isDeploying ? 'animate-pulse-fast' : ''}`}></span>
      {showText && <span className="text-sm text-gray-300">{text}</span>}
    </div>
  );
};

export default StatusIndicator;
