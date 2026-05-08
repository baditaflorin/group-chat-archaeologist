import { useQuery } from '@tanstack/react-query';
import { assetUrl } from './api';

type BuildInfo = {
  version: string;
  commit: string;
};

export function useBuildInfo() {
  return useQuery({
    queryKey: ['build-info'],
    queryFn: async (): Promise<BuildInfo> => {
      const response = await fetch(assetUrl('build-info.json'));
      if (!response.ok) {
        throw new Error(`Build info lookup failed: ${response.status}`);
      }
      const text = await response.text();
      const jsonStart = text.indexOf('{');
      const parsed = JSON.parse(text.slice(jsonStart)) as BuildInfo;
      if (!parsed.commit || parsed.commit.includes('{{')) {
        throw new Error('Build info is not rendered yet');
      }
      return {
        version: parsed.version,
        commit: parsed.commit.slice(0, 7)
      };
    },
    retry: false,
    staleTime: Number.POSITIVE_INFINITY
  });
}
