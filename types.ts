
export enum ApplicationStatus {
  HEALTHY = 'Healthy',
  UNHEALTHY = 'Unhealthy',
  DEPLOYING = 'Deploying',
  STOPPED = 'Stopped',
}

export interface Metric {
  time: number;
  value: number;
}

export interface Application {
  id: string;
  name: string;
  dockerImage: string;
  status: ApplicationStatus;
  replicas: {
    current: number;
    target: number;
  };
  cpuUsage: Metric[];
  memoryUsage: Metric[];
  logs: string[];
  createdAt: number;
}
