
import React from 'react';
import type { Application } from '../types';
import StatusIndicator from './icons/StatusIndicator';
import { PlusIcon, MinusIcon, TrashIcon, ExternalLinkIcon } from './icons/Icon';

interface AppCardProps {
  app: Application;
  onScale: (appId: string, newTargetReplicas: number) => void;
  onRemove: (appId: string) => void;
  onViewDetails: (app: Application) => void;
}

const AppCard: React.FC<AppCardProps> = ({ app, onScale, onRemove, onViewDetails }) => {

  const handleScaleUp = () => {
    onScale(app.id, app.replicas.target + 1);
  };

  const handleScaleDown = () => {
    onScale(app.id, app.replicas.target - 1);
  };
  
  const lastCpu = app.cpuUsage[app.cpuUsage.length - 1]?.value ?? 0;
  const lastMem = app.memoryUsage[app.memoryUsage.length - 1]?.value ?? 0;

  return (
    <div className="bg-gray-800 rounded-lg shadow-lg overflow-hidden flex flex-col transition-all duration-300 hover:shadow-2xl hover:-translate-y-1">
      <div className="p-5 flex-grow">
        <div className="flex justify-between items-start">
            <div className="flex-1 min-w-0">
                <h3 className="text-lg font-bold text-gray-200 truncate">{app.name}</h3>
                <p className="text-sm text-gray-500 truncate">{app.dockerImage}</p>
            </div>
            <StatusIndicator status={app.status} />
        </div>
        
        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
            <div>
                <span className="text-gray-500">CPU</span>
                <p className="text-gray-200 font-semibold">{lastCpu.toFixed(1)}%</p>
            </div>
            <div>
                <span className="text-gray-500">Memory</span>
                <p className="text-gray-200 font-semibold">{lastMem.toFixed(0)} MB</p>
            </div>
        </div>
        
        <div className="mt-4">
            <span className="text-gray-500 text-sm">Replicas</span>
            <div className="flex items-center space-x-3 mt-1">
                 <button onClick={handleScaleDown} className="p-1 rounded-full bg-gray-700 hover:bg-gray-600 disabled:opacity-50" disabled={app.replicas.target <= 0}>
                    <MinusIcon className="w-4 h-4 text-gray-300" />
                </button>
                <span className="font-semibold text-gray-200 text-lg w-12 text-center">{app.replicas.current} / {app.replicas.target}</span>
                <button onClick={handleScaleUp} className="p-1 rounded-full bg-gray-700 hover:bg-gray-600">
                    <PlusIcon className="w-4 h-4 text-gray-300" />
                </button>
            </div>
        </div>
      </div>
      
      <div className="bg-gray-700/50 px-5 py-3 flex items-center justify-between">
         <button onClick={() => onViewDetails(app)} className="inline-flex items-center text-sm text-blue-400 hover:text-blue-300 font-semibold">
            <ExternalLinkIcon className="w-4 h-4 mr-2" />
            View Details
        </button>
        <button onClick={() => onRemove(app.id)} className="p-2 rounded-full hover:bg-red-500/20 text-gray-500 hover:text-red-400">
            <TrashIcon className="w-4 h-4"/>
        </button>
      </div>
    </div>
  );
};

export default AppCard;
