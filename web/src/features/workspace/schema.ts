import { z } from 'zod';
import { dashboardSchema } from '../chat/schema';

export const viewSchema = z.enum(['timeline', 'map', 'jokes', 'departures', 'settings']);

export const workspaceSettingsSchema = z.object({
  persistWorkspace: z.boolean(),
  showWarnings: z.boolean(),
  showConfidence: z.boolean(),
  preferRestore: z.boolean()
});

export const importRecordSchema = z.object({
  name: z.string(),
  kind: z.enum(['demo', 'raw', 'dashboard', 'workspace']),
  status: z.enum(['loaded', 'skipped', 'failed']),
  adapter: z.string().optional(),
  messages: z.number().optional(),
  warnings: z.number().optional(),
  detail: z.string().optional()
});

export const workspaceStateSchema = z.object({
  schemaVersion: z.literal('v1'),
  savedAt: z.string(),
  source: z.object({
    kind: z.enum(['demo', 'raw', 'dashboard', 'workspace', 'shared']),
    label: z.string(),
    note: z.string().optional()
  }),
  dashboard: dashboardSchema,
  search: z.string(),
  selectedMembers: z.array(z.string()),
  activeView: viewSchema,
  settings: workspaceSettingsSchema,
  imports: z.array(importRecordSchema)
});

export type View = z.infer<typeof viewSchema>;
export type WorkspaceSettings = z.infer<typeof workspaceSettingsSchema>;
export type ImportRecord = z.infer<typeof importRecordSchema>;
export type WorkspaceState = z.infer<typeof workspaceStateSchema>;

export const defaultWorkspaceSettings: WorkspaceSettings = {
  persistWorkspace: true,
  showWarnings: true,
  showConfidence: true,
  preferRestore: true
};
