import axios from 'axios';
import { Application, EnvVar, PortMapping } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export const apiService = {
  // Get all applications
  getApplications: async (): Promise<Application[]> => {
    const response = await axios.get(`${API_BASE_URL}/applications`);
    // Ensure we always return an array even if the server returns null initially
    return response.data || [];
  },

  // Deploy a new application
  deployApplication: async (
    image: string, 
    replicas: number, 
    envVars: EnvVar[] = [], 
    ports: PortMapping[] = []
  ): Promise<Application> => {
    const payload = {
        image,
        replicas,
        envVars,
        ports
    };
    const response = await axios.post(`${API_BASE_URL}/applications`, payload);
    return response.data;
  },

  // Scale application
  scaleApplication: async (appId: string, targetReplicas: number): Promise<Application> => {
    const response = await axios.patch(`${API_BASE_URL}/applications?id=${appId}`, {
        replicas: targetReplicas
    });
    return response.data;
  },

  // Remove application
  removeApplication: async (appId: string): Promise<void> => {
    await axios.delete(`${API_BASE_URL}/applications?id=${appId}`);
  }
};
