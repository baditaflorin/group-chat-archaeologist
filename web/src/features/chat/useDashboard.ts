import { useQuery } from '@tanstack/react-query';
import { fetchDashboard, fetchMeta } from './api';

export function useDashboard() {
  return useQuery({
    queryKey: ['chat-archaeology-dashboard', 'v1'],
    queryFn: fetchDashboard
  });
}

export function useMeta() {
  return useQuery({
    queryKey: ['chat-archaeology-meta', 'v1'],
    queryFn: fetchMeta
  });
}
