import { dashboardSchema, metaSchema, type Dashboard, type Meta } from './schema';

export function assetUrl(path: string) {
  return `${import.meta.env.BASE_URL}${path}`;
}

async function getJson<T>(path: string, parse: (value: unknown) => T): Promise<T> {
  const response = await fetch(assetUrl(path));
  if (!response.ok) {
    throw new Error(`Fetch failed for ${path}: ${response.status}`);
  }
  return parse(await response.json());
}

export function fetchDashboard(): Promise<Dashboard> {
  return getJson('data/v1/chat-archaeology.json', (value) => dashboardSchema.parse(value));
}

export function fetchMeta(): Promise<Meta> {
  return getJson('data/v1/chat-archaeology.meta.json', (value) => metaSchema.parse(value));
}
