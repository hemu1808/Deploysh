
import React, { useState } from 'react';

interface DeployAppModalProps {
  onClose: () => void;
  onDeploy: (image: string, replicas: number) => void;
}

const DeployAppModal: React.FC<DeployAppModalProps> = ({ onClose, onDeploy }) => {
  const [image, setImage] = useState('');
  const [replicas, setReplicas] = useState(1);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (image.trim()) {
      onDeploy(image.trim(), replicas);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 transition-opacity" onClick={onClose}>
      <div className="bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-bold text-gray-200 mb-4">Deploy New Application</h2>
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="docker-image" className="block text-sm font-medium text-gray-400 mb-1">
              Docker Image
            </label>
            <input
              type="text"
              id="docker-image"
              value={image}
              onChange={(e) => setImage(e.target.value)}
              className="block w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md shadow-sm placeholder-gray-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm text-gray-200"
              placeholder="e.g., nginx:latest"
              required
              autoFocus
            />
          </div>
          <div className="mb-6">
            <label htmlFor="replicas" className="block text-sm font-medium text-gray-400 mb-1">
              Replicas
            </label>
            <input
              type="number"
              id="replicas"
              value={replicas}
              onChange={(e) => setReplicas(Math.max(1, parseInt(e.target.value, 10)))}
              className="block w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm text-gray-200"
              min="1"
            />
          </div>
          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-300 bg-gray-700 rounded-md hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-800 focus:ring-gray-500"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-500 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-800 focus:ring-blue-500 disabled:opacity-50"
              disabled={!image.trim()}
            >
              Deploy
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default DeployAppModal;
