
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
    <div className="glass-card rounded-xl overflow-hidden flex flex-col transition-all duration-300 hover:-translate-y-2 hover:shadow-[0_8px_30px_rgba(56,189,248,0.15)] group relative">
      <div className="absolute inset-0 bg-gradient-to-br from-brand-blue/5 to-brand-purple/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"></div>
      <div className="p-5 flex-grow">
        <div className="flex justify-between items-start">
            <div className="flex-1 min-w-0">
                <h3 className="text-lg font-bold text-slate-100 truncate tracking-wide">{app.name}</h3>
                <p className="text-xs text-brand-blue/80 font-mono truncate mt-1 bg-brand-blue/10 px-2 py-0.5 rounded-full inline-block">{app.dockerImage}</p>
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
        
        <div className="mt-5 border-t border-glass-border pt-4">
            <span className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Replicas</span>
            <div className="flex items-center space-x-4 mt-2">
                 <button onClick={handleScaleDown} className="p-1.5 rounded-full bg-slate-800/50 border border-glass-border hover:bg-slate-700/80 hover:border-brand-blue/30 transition-all disabled:opacity-30 disabled:hover:border-glass-border" disabled={app.replicas.target <= 0}>
                    <MinusIcon className="w-3.5 h-3.5 text-slate-300" />
                </button>
                <div className="flex flex-col items-center justify-center w-14">
                  <span className="font-bold text-slate-100 text-xl leading-none">{app.replicas.current}</span>
                  <span className="text-[10px] text-slate-500 font-mono mt-0.5">/ {app.replicas.target} TGT</span>
                </div>
                <button onClick={handleScaleUp} className="p-1.5 rounded-full bg-slate-800/50 border border-glass-border hover:bg-slate-700/80 hover:border-brand-blue/30 transition-all">
                    <PlusIcon className="w-3.5 h-3.5 text-slate-300" />
                </button>
            </div>
        </div>
      </div>
      
      <div className="glass-panel border-x-0 border-b-0 px-5 py-3.5 flex items-center justify-between z-10">
         <button onClick={() => onViewDetails(app)} className="inline-flex items-center text-sm text-brand-blue hover:text-brand-purple transition-colors font-semibold">
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
