
import React, { useEffect, useRef } from 'react';

interface LogViewerProps {
  logs: string[];
}

const LogViewer: React.FC<LogViewerProps> = ({ logs }) => {
  const logContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs]);

  const renderLogLine = (line: string, index: number) => {
    let colorClass = 'text-gray-400';
    if (line.includes('[ERROR]')) colorClass = 'text-red-400';
    else if (line.includes('[WARN]')) colorClass = 'text-yellow-400';
    else if (line.includes('[INFO]')) colorClass = 'text-blue-400';
    else if (line.includes('[DEBUG]')) colorClass = 'text-purple-400';

    return (
        <div key={index} className="flex">
            <span className="text-gray-600 mr-4 select-none w-8 text-right">{index + 1}</span>
            <span className={colorClass}>{line}</span>
        </div>
    );
  };

  return (
    <div ref={logContainerRef} className="bg-gray-900 text-sm font-mono p-4 rounded-md h-full overflow-y-auto">
        {logs.map(renderLogLine)}
    </div>
  );
};

export default LogViewer;
