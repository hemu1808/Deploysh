
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

export interface EnvVar {
  key: string;
  value: string;
}

export interface PortMapping {
  hostPort: number;
  containerPort: number;
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
  envVars?: EnvVar[];
  ports?: PortMapping[];
  createdAt: number;
}
