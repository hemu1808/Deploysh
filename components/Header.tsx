
import React from 'react';
import { RocketIcon } from './icons/Icon';

interface HeaderProps {
  onDeployClick: () => void;
}

const Header: React.FC<HeaderProps> = ({ onDeployClick }) => {
  return (
    <header className="bg-gray-800/50 backdrop-blur-sm sticky top-0 z-10 border-b border-gray-700">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-3">
            <RocketIcon className="h-8 w-8 text-blue-500" />
            <h1 className="text-xl font-bold text-gray-200">AuraDeploy</h1>
          </div>
          <button
            onClick={onDeployClick}
            className="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-500 hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-900 focus:ring-blue-500 transition-colors"
          >
            Deploy New App
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;
