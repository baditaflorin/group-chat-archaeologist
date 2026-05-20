export async function fetchText(path: string): Promise<string> {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`Fetch failed for ${path}: ${response.status}`);
  }
  return response.text();
}

export async function fetchJson<T>(path: string, parse: (value: unknown) => T): Promise<T> {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`Fetch failed for ${path}: ${response.status}`);
  }
  return parse(await response.json());
}
