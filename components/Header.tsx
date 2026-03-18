
import React from 'react';
import { RocketIcon } from './icons/Icon';

interface HeaderProps {
  onDeployClick: () => void;
}

const Header: React.FC<HeaderProps> = ({ onDeployClick }) => {
  return (
    <header className="glass-panel sticky top-0 z-10 border-b border-glass-border">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-3">
            <RocketIcon className="h-8 w-8 text-blue-500" />
            <h1 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-brand-blue to-brand-purple tracking-tight">AuraDeploy</h1>
          </div>
          <button
            onClick={onDeployClick}
            className="inline-flex items-center justify-center px-5 py-2.5 border border-brand-blue/30 text-sm font-semibold rounded-lg shadow-[0_0_15px_rgba(56,189,248,0.3)] text-white bg-brand-blue/10 backdrop-blur-sm hover:bg-brand-blue/20 hover:shadow-[0_0_25px_rgba(56,189,248,0.5)] focus:outline-none focus:ring-2 focus:ring-brand-blue transition-all duration-300"
          >
            Deploy New App
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;
