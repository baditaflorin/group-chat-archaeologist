import { describe, expect, it } from 'vitest';
import { dashboardSchema } from './schema';

describe('dashboard schema', () => {
  it('validates the generated artifact shape', async () => {
    const artifact = await import('../../../../docs/data/v1/chat-archaeology.json');
    expect(() => dashboardSchema.parse(artifact.default)).not.toThrow();
  });
});
