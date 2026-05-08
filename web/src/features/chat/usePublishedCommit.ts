import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';

const commitSchema = z.object({
  sha: z.string()
});

export function usePublishedCommit() {
  return useQuery({
    queryKey: ['published-commit', 'main'],
    queryFn: async () => {
      const response = await fetch(
        'https://api.github.com/repos/baditaflorin/group-chat-archaeologist/commits/main'
      );
      if (!response.ok) {
        throw new Error(`GitHub commit lookup failed: ${response.status}`);
      }
      return commitSchema.parse(await response.json()).sha.slice(0, 7);
    },
    retry: false,
    staleTime: 1000 * 60 * 10
  });
}
