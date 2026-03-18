import React, { useState } from 'react';
import { EnvVar, PortMapping } from '../types';
import { PlusIcon, MinusIcon } from './icons/Icon';

interface DeployAppModalProps {
  onClose: () => void;
  onDeploy: (image: string, replicas: number, envVars: EnvVar[], ports: PortMapping[]) => void;
}

const DeployAppModal: React.FC<DeployAppModalProps> = ({ onClose, onDeploy }) => {
  const [image, setImage] = useState('');
  const [replicas, setReplicas] = useState(1);
  const [envVars, setEnvVars] = useState<EnvVar[]>([]);
  const [ports, setPorts] = useState<PortMapping[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (image.trim()) {
      onDeploy(
          image.trim(), 
          replicas, 
          envVars.filter(e => e.key.trim() !== ''), 
          ports.filter(p => p.hostPort > 0 && p.containerPort > 0)
      );
    }
  };

  const addEnvVar = () => setEnvVars([...envVars, { key: '', value: '' }]);
  const removeEnvVar = (index: number) => setEnvVars(envVars.filter((_, i) => i !== index));
  const updateEnvVar = (index: number, field: 'key' | 'value', val: string) => {
      const newVars = [...envVars];
      newVars[index][field] = val;
      setEnvVars(newVars);
  };

  const addPort = () => setPorts([...ports, { hostPort: 0, containerPort: 0 }]);
  const removePort = (index: number) => setPorts(ports.filter((_, i) => i !== index));
  const updatePort = (index: number, field: 'hostPort' | 'containerPort', val: number) => {
      const newPorts = [...ports];
      newPorts[index][field] = val;
      setPorts(newPorts);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 transition-opacity" onClick={onClose}>
      <div className="bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 p-6 max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-bold text-gray-200 mb-6 border-b border-gray-700 pb-2">Deploy New Application</h2>
        
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Docker Image</label>
                <input
                  type="text"
                  value={image}
                  onChange={(e) => setImage(e.target.value)}
                  className="block w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-200 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="e.g., nginx:latest"
                  required
                  autoFocus
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Replicas</label>
                <input
                  type="number"
                  value={replicas}
                  onChange={(e) => setReplicas(Math.max(1, parseInt(e.target.value, 10)))}
                  className="block w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-200 focus:ring-blue-500 focus:border-blue-500"
                  min="1"
                />
              </div>
          </div>

          {/* Environment Variables Section */}
          <div>
            <div className="flex justify-between items-center mb-2">
                <label className="block text-sm font-medium text-gray-400">Environment Variables</label>
                <button type="button" onClick={addEnvVar} className="text-xs text-blue-400 hover:text-blue-300 flex items-center">
                    <PlusIcon className="w-3 h-3 mr-1" /> Add Variable
                </button>
            </div>
            {envVars.length === 0 && <p className="text-xs text-gray-500 italic">No environment variables defined.</p>}
            <div className="space-y-2 max-h-40 overflow-y-auto pr-2">
                {envVars.map((env, index) => (
                    <div key={index} className="flex space-x-2">
                        <input
                            type="text"
                            placeholder="Key (e.g., DB_HOST)"
                            value={env.key}
                            onChange={(e) => updateEnvVar(index, 'key', e.target.value)}
                            className="flex-1 px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-200 focus:ring-blue-500"
                        />
                        <input
                            type="text"
                            placeholder="Value"
                            value={env.value}
                            onChange={(e) => updateEnvVar(index, 'value', e.target.value)}
                            className="flex-1 px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-200 focus:ring-blue-500"
                        />
                        <button type="button" onClick={() => removeEnvVar(index)} className="p-1.5 text-gray-500 hover:text-red-400">
                            <MinusIcon className="w-4 h-4" />
                        </button>
                    </div>
                ))}
            </div>
          </div>

          {/* Port Mappings Section */}
          <div>
            <div className="flex justify-between items-center mb-2">
                <label className="block text-sm font-medium text-gray-400">Port Configuration</label>
                <button type="button" onClick={addPort} className="text-xs text-blue-400 hover:text-blue-300 flex items-center">
                    <PlusIcon className="w-3 h-3 mr-1" /> Add Port
                </button>
            </div>
            {ports.length === 0 && <p className="text-xs text-gray-500 italic">No ports exposed.</p>}
            <div className="space-y-2 max-h-40 overflow-y-auto pr-2">
                {ports.map((port, index) => (
                    <div key={index} className="flex space-x-2 items-center">
                         <span className="text-sm text-gray-500">Host:</span>
                        <input
                            type="number"
                            placeholder="8080"
                            value={port.hostPort || ''}
                            onChange={(e) => updatePort(index, 'hostPort', parseInt(e.target.value, 10))}
                            className="w-24 px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-200"
                        />
                        <span className="text-sm text-gray-500">-&gt; Container:</span>
                        <input
                            type="number"
                            placeholder="80"
                            value={port.containerPort || ''}
                            onChange={(e) => updatePort(index, 'containerPort', parseInt(e.target.value, 10))}
                            className="w-24 px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-200"
                        />
                        <button type="button" onClick={() => removePort(index)} className="p-1.5 text-gray-500 hover:text-red-400 ml-auto">
                            <MinusIcon className="w-4 h-4" />
                        </button>
                    </div>
                ))}
            </div>
          </div>

          <div className="flex justify-end space-x-3 pt-4 border-t border-gray-700">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-300 bg-gray-700 rounded-md hover:bg-gray-600 focus:outline-none"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-500 rounded-md hover:bg-blue-600 focus:outline-none disabled:opacity-50"
              disabled={!image.trim()}
            >
              Deploy Service
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default DeployAppModal;
