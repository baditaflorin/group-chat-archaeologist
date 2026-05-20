import { fetchJson } from './fetch';
import { dashboardSchema, metaSchema, type Dashboard, type Meta } from './schema';

export function assetUrl(path: string) {
  return `${import.meta.env.BASE_URL}${path}`;
}

export function fetchDashboard(): Promise<Dashboard> {
  return fetchJson(assetUrl('data/v1/chat-archaeology.json'), (value) => dashboardSchema.parse(value));
}

export function fetchMeta(): Promise<Meta> {
  return fetchJson(assetUrl('data/v1/chat-archaeology.meta.json'), (value) => metaSchema.parse(value));
}
