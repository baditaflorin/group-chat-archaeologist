import { workspaceStateSchema, type WorkspaceState } from './schema';

const storageKey = 'group-chat-archaeologist.workspace.v1';

export function loadStoredWorkspace(): WorkspaceState | null {
  const raw = localStorage.getItem(storageKey);
  if (!raw) {
    return null;
  }
  try {
    return workspaceStateSchema.parse(JSON.parse(raw));
  } catch {
    return null;
  }
}

export function saveStoredWorkspace(workspace: WorkspaceState) {
  localStorage.setItem(storageKey, JSON.stringify(workspace));
}

export function clearStoredWorkspace() {
  localStorage.removeItem(storageKey);
}

export function parseSharedWorkspaceHash(hash: string): WorkspaceState | null {
  if (!hash.startsWith('#state=')) {
    return null;
  }
  try {
    const encoded = hash.slice('#state='.length);
    const json = decodeURIComponent(atob(encoded));
    return workspaceStateSchema.parse(JSON.parse(json));
  } catch {
    return null;
  }
}

export function buildSharedWorkspaceHash(workspace: WorkspaceState): string | null {
  const json = JSON.stringify(workspace);
  const encoded = btoa(encodeURIComponent(json));
  const hash = `#state=${encoded}`;
  if (hash.length > 3800) {
    return null;
  }
  return hash;
}
