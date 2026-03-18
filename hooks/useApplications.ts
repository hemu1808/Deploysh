import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { apiService } from '../services/api';
import { wsService } from '../services/websocket';
import { Application, EnvVar, PortMapping } from '../types';

export const useApplications = () => {
  const queryClient = useQueryClient();

  // Fetch initial data via REST
  const { data: applications = [], isLoading, error } = useQuery({
    queryKey: ['applications'],
    queryFn: apiService.getApplications,
  });

  // Setup WebSocket connection to receive real-time updates seamlessly merging into the React Query cache
  useEffect(() => {
    let unsubscribe: (() => void) | undefined;
    
    wsService.connect(() => {
        unsubscribe = wsService.subscribe((msg: any) => {
            if (msg.type === 'SYNC_APPS') {
                queryClient.setQueryData(['applications'], msg.payload as Application[]);
            }
        });
    });

    return () => {
        if (unsubscribe) unsubscribe();
        wsService.disconnect();
    };
  }, [queryClient]);

  // Mutations
  const deployMutation = useMutation({
    mutationFn: (vars: { image: string, replicas: number, envVars?: EnvVar[], ports?: PortMapping[] }) => 
        apiService.deployApplication(vars.image, vars.replicas, vars.envVars, vars.ports),
    onSuccess: () => {
        // Invalidate and refetch
        queryClient.invalidateQueries({ queryKey: ['applications'] });
    },
  });

  const scaleMutation = useMutation({
    mutationFn: (vars: { appId: string, replicas: number }) => 
        apiService.scaleApplication(vars.appId, vars.replicas),
    onSuccess: () => {
        // Or optimally optimistic update
        queryClient.invalidateQueries({ queryKey: ['applications'] });
    },
  });

  const removeMutation = useMutation({
    mutationFn: (appId: string) => apiService.removeApplication(appId),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['applications'] });
    },
  });

  return {
    applications,
    isLoading,
    error,
    deployApplication: deployMutation.mutate,
    isDeploying: deployMutation.isPending,
    scaleApplication: scaleMutation.mutate,
    isScaling: scaleMutation.isPending,
    removeApplication: removeMutation.mutate,
    isRemoving: removeMutation.isPending
  };
};
