
import { useState, useEffect, useCallback } from 'react';
import type { Application, Metric } from '../types';
import { ApplicationStatus } from '../types';

const MAX_METRIC_POINTS = 30;
const MAX_LOG_LINES = 100;

const initialApps: Application[] = [
  {
    id: 'app-1',
    name: 'api-gateway',
    dockerImage: 'nginx:latest',
    status: ApplicationStatus.HEALTHY,
    replicas: { current: 3, target: 3 },
    cpuUsage: Array.from({ length: MAX_METRIC_POINTS }, (_, i) => ({ time: Date.now() - (MAX_METRIC_POINTS - i) * 5000, value: Math.random() * 15 + 5 })),
    memoryUsage: Array.from({ length: MAX_METRIC_POINTS }, (_, i) => ({ time: Date.now() - (MAX_METRIC_POINTS - i) * 5000, value: Math.random() * 200 + 100 })),
    logs: ['[INFO] GET /api/v1/users 200 OK', '[INFO] Service started successfully on port 80.'],
    createdAt: Date.now() - 3600000,
  },
  {
    id: 'app-2',
    name: 'user-service',
    dockerImage: 'myapp/user-service:1.2.0',
    status: ApplicationStatus.UNHEALTHY,
    replicas: { current: 1, target: 2 },
    cpuUsage: Array.from({ length: MAX_METRIC_POINTS }, (_, i) => ({ time: Date.now() - (MAX_METRIC_POINTS - i) * 5000, value: Math.random() * 50 + 40 })),
    memoryUsage: Array.from({ length: MAX_METRIC_POINTS }, (_, i) => ({ time: Date.now() - (MAX_METRIC_POINTS - i) * 5000, value: Math.random() * 100 + 400 })),
    logs: ['[ERROR] Database connection failed: timeout expired.', '[WARN] High memory usage detected: 480MB'],
    createdAt: Date.now() - 7200000,
  },
    {
    id: 'app-3',
    name: 'frontend-webapp',
    dockerImage: 'node:18-alpine',
    status: ApplicationStatus.DEPLOYING,
    replicas: { current: 0, target: 2 },
    cpuUsage: [],
    memoryUsage: [],
    logs: ['[INFO] Starting deployment...', '[INFO] Pulling docker image node:18-alpine...'],
    createdAt: Date.now() - 60000,
  },
];

const generateLogLine = (appName: string): string => {
    const levels = ['INFO', 'WARN', 'ERROR', 'DEBUG'];
    const methods = ['GET', 'POST', 'PUT', 'DELETE'];
    const paths = ['/api/v1/items', '/api/v1/users', '/healthz', '/metrics'];
    const statuses = [200, 201, 400, 404, 500];
    const messages = ['Request processed', 'User authenticated', 'Item not found', 'Internal server error'];
    
    const level = levels[Math.floor(Math.random() * levels.length)];
    const message = `[${level}] ${methods[Math.floor(Math.random() * methods.length)]} ${paths[Math.floor(Math.random() * paths.length)]} ${statuses[Math.floor(Math.random() * statuses.length)]} ${messages[Math.floor(Math.random() * messages.length)]}`;
    return message;
};


export const useMockData = () => {
  const [applications, setApplications] = useState<Application[]>(initialApps);

  useEffect(() => {
    const interval = setInterval(() => {
      setApplications(prevApps =>
        prevApps.map(app => {
          // Update status
          let newStatus = app.status;
          if (app.status === ApplicationStatus.DEPLOYING) {
             if (Math.random() > 0.7) {
                 newStatus = ApplicationStatus.HEALTHY;
                 app.replicas.current = app.replicas.target;
             }
          } else if (Math.random() > 0.98) {
            newStatus = newStatus === ApplicationStatus.HEALTHY ? ApplicationStatus.UNHEALTHY : ApplicationStatus.HEALTHY;
          }

          // Update metrics
          const newCpuMetric: Metric = { time: Date.now(), value: newStatus === ApplicationStatus.UNHEALTHY ? Math.random() * 50 + 40 : Math.random() * 15 + 5 };
          const newMemoryMetric: Metric = { time: Date.now(), value: newStatus === ApplicationStatus.UNHEALTHY ? Math.random() * 100 + 400 : Math.random() * 200 + 100 };
          
          const cpuUsage = [...app.cpuUsage, newCpuMetric].slice(-MAX_METRIC_POINTS);
          const memoryUsage = [...app.memoryUsage, newMemoryMetric].slice(-MAX_METRIC_POINTS);
          
          // Update logs
          const newLog = generateLogLine(app.name);
          const logs = [...app.logs, newLog].slice(-MAX_LOG_LINES);
          
          return { ...app, status: newStatus, cpuUsage, memoryUsage, logs };
        })
      );
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  const addApplication = useCallback((image: string, replicas: number) => {
    const name = image.split(':')[0].split('/').pop() || `app-${Math.floor(Math.random() * 1000)}`;
    const newApp: Application = {
      id: `app-${Date.now()}`,
      name: name,
      dockerImage: image,
      status: ApplicationStatus.DEPLOYING,
      replicas: { current: 0, target: replicas },
      cpuUsage: [],
      memoryUsage: [],
      logs: ['[INFO] Deployment initiated by user.'],
      createdAt: Date.now(),
    };
    setApplications(prev => [...prev, newApp]);
  }, []);

  const removeApplication = useCallback((appId: string) => {
    setApplications(prev => prev.filter(app => app.id !== appId));
  }, []);
  
  const scaleApplication = useCallback((appId: string, newTargetReplicas: number) => {
    setApplications(prev => prev.map(app => 
      app.id === appId 
        ? { ...app, replicas: { ...app.replicas, target: Math.max(0, newTargetReplicas) }, status: ApplicationStatus.DEPLOYING }
        : app
    ));
  }, []);

  return { applications, addApplication, removeApplication, scaleApplication };
};
