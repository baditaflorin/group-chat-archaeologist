import { useEffect, useState } from 'react';
import type { Dashboard } from '../chat/schema';
import { buildSummaryText, importFromFiles, importFromPastedText } from './browser-import';
import {
  buildSharedWorkspaceHash,
  clearStoredWorkspace,
  loadStoredWorkspace,
  parseSharedWorkspaceHash,
  saveStoredWorkspace
} from './storage';
import { defaultWorkspaceSettings, type View, type WorkspaceState } from './schema';

type Notice = {
  tone: 'neutral' | 'success' | 'warning';
  body: string;
};

function makeDemoWorkspace(dashboard: Dashboard): WorkspaceState {
  return {
    schemaVersion: 'v1',
    savedAt: new Date().toISOString(),
    source: {
      kind: 'demo',
      label: 'Shipped demo archive',
      note: 'This is the built-in sample artifact.'
    },
    dashboard,
    search: '',
    selectedMembers: [],
    activeView: 'timeline',
    settings: defaultWorkspaceSettings,
    imports: [
      {
        name: 'demo',
        kind: 'demo',
        status: 'loaded',
        adapter: dashboard.source.adapter ?? dashboard.source.parser,
        messages: dashboard.source.messageCount,
        warnings: dashboard.source.warningCount ?? dashboard.warnings?.length ?? 0,
        detail: 'Loaded the shipped demo artifact.'
      }
    ]
  };
}

export function useWorkspace(demoDashboard: Dashboard | undefined) {
  const [workspace, setWorkspace] = useState<WorkspaceState | null>(() => {
    const fromHash = parseSharedWorkspaceHash(window.location.hash);
    if (fromHash) return fromHash;
    const stored = loadStoredWorkspace();
    if (stored?.settings.preferRestore) return stored;
    return null;
  });
  const [notice, setNotice] = useState<Notice | null>(() => {
    if (parseSharedWorkspaceHash(window.location.hash)) {
      return { tone: 'success', body: 'Loaded a shared workspace from the URL.' };
    }
    const stored = loadStoredWorkspace();
    if (stored?.settings.preferRestore) {
      return { tone: 'success', body: 'Restored your last workspace from this browser.' };
    }
    return null;
  });
  const [busy, setBusy] = useState(false);

  // Lazy demo fallback: only fires when demoDashboard first loads and nothing
  // was restored from the URL hash or local storage.
  useEffect(() => {
    if (workspace !== null || !demoDashboard) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setWorkspace(makeDemoWorkspace(demoDashboard));
  }, [demoDashboard, workspace]);

  useEffect(() => {
    if (!workspace) {
      return;
    }
    if (workspace.settings.persistWorkspace) {
      saveStoredWorkspace(workspace);
    } else {
      clearStoredWorkspace();
    }
  }, [workspace]);

  function updateWorkspace(update: (current: WorkspaceState) => WorkspaceState) {
    setWorkspace((current) => {
      if (!current) {
        return current;
      }
      return {
        ...update(current),
        savedAt: new Date().toISOString()
      };
    });
  }

  function loadDemo() {
    if (!demoDashboard) {
      return;
    }
    setWorkspace(makeDemoWorkspace(demoDashboard));
    setNotice({ tone: 'success', body: 'Switched back to the shipped demo archive.' });
  }

  async function importFiles(files: FileList | File[]) {
    if (!files.length) {
      return;
    }
    setBusy(true);
    try {
      const imported = await importFromFiles(files);
      setWorkspace(imported.workspace);
      setNotice({
        tone: 'success',
        body: `Loaded ${imported.records.filter((record) => record.status === 'loaded').length} file(s) into this workspace.`
      });
    } catch (error) {
      setNotice({
        tone: 'warning',
        body: error instanceof Error ? error.message : 'The selected files could not be imported.'
      });
    } finally {
      setBusy(false);
    }
  }

  async function importPastedValue(name: string, value: string, extractionMode: string) {
    if (!value.trim()) {
      setNotice({ tone: 'warning', body: 'Paste a chat export, dashboard JSON, or saved workspace first.' });
      return;
    }
    setBusy(true);
    try {
      const imported = await importFromPastedText(name, value, extractionMode);
      setWorkspace(imported.workspace);
      setNotice({ tone: 'success', body: `Loaded ${name} into the browser workspace.` });
    } catch (error) {
      setNotice({
        tone: 'warning',
        body: error instanceof Error ? error.message : 'The pasted content could not be imported.'
      });
    } finally {
      setBusy(false);
    }
  }

  async function importClipboard() {
    if (!navigator.clipboard?.readText) {
      setNotice({
        tone: 'warning',
        body: 'Clipboard read is not available in this browser. Paste into the text box instead.'
      });
      return;
    }
    try {
      const text = await navigator.clipboard.readText();
      await importPastedValue('clipboard import', text, 'txt');
    } catch {
      setNotice({ tone: 'warning', body: 'Clipboard access was blocked. Paste into the text box instead.' });
    }
  }

  function clearWorkspace() {
    clearStoredWorkspace();
    if (demoDashboard) {
      setWorkspace(makeDemoWorkspace(demoDashboard));
    }
    window.location.hash = '';
    setNotice({ tone: 'success', body: 'Started fresh and cleared this browser workspace.' });
  }

  function setActiveView(view: View) {
    updateWorkspace((current) => ({ ...current, activeView: view }));
  }

  function setSearch(value: string) {
    updateWorkspace((current) => ({ ...current, search: value }));
  }

  function toggleMember(member: string) {
    updateWorkspace((current) => {
      const selected = current.selectedMembers.includes(member)
        ? current.selectedMembers.filter((item) => item !== member)
        : [...current.selectedMembers, member];
      return { ...current, selectedMembers: selected };
    });
  }

  function resetMembers() {
    updateWorkspace((current) => ({ ...current, selectedMembers: [] }));
  }

  function updateSetting<Key extends keyof WorkspaceState['settings']>(
    key: Key,
    value: WorkspaceState['settings'][Key]
  ) {
    updateWorkspace((current) => ({
      ...current,
      settings: {
        ...current.settings,
        [key]: value
      }
    }));
  }

  async function copySummary() {
    if (!workspace) {
      return;
    }
    try {
      await navigator.clipboard.writeText(buildSummaryText(workspace.dashboard));
      setNotice({ tone: 'success', body: 'Copied a human-readable archive summary.' });
    } catch {
      setNotice({ tone: 'warning', body: 'Clipboard write failed in this browser.' });
    }
  }

  async function shareWorkspace() {
    if (!workspace) {
      return;
    }
    const hash = buildSharedWorkspaceHash(workspace);
    if (!hash) {
      setNotice({
        tone: 'warning',
        body: 'This workspace is too large for a shareable URL. Use Save Workspace instead.'
      });
      return;
    }
    const url = `${window.location.origin}${window.location.pathname}${hash}`;
    window.location.hash = hash;
    try {
      await navigator.clipboard.writeText(url);
      setNotice({ tone: 'success', body: 'Copied a shareable workspace URL.' });
    } catch {
      setNotice({ tone: 'success', body: 'Updated the URL hash with this workspace state.' });
    }
  }

  function printWorkspace() {
    window.print();
  }

  function downloadDashboard() {
    if (!workspace) {
      return;
    }
    downloadJson('chat-archaeology.json', workspace.dashboard);
    setNotice({ tone: 'success', body: 'Downloaded the current dashboard JSON.' });
  }

  function downloadWorkspace() {
    if (!workspace) {
      return;
    }
    downloadJson('group-chat-archaeologist-workspace.json', workspace);
    setNotice({ tone: 'success', body: 'Downloaded the saved browser workspace.' });
  }

  function downloadGraph() {
    if (!workspace) {
      return;
    }
    const { svgPath } = workspace.dashboard.graph;
    if (!svgPath.startsWith('data:image/svg+xml')) {
      window.open(svgPath, '_blank', 'noopener,noreferrer');
      return;
    }
    const anchor = document.createElement('a');
    anchor.href = svgPath;
    anchor.download = 'who-introduced-whom.svg';
    anchor.click();
    setNotice({ tone: 'success', body: 'Downloaded the relationship map SVG.' });
  }

  return {
    workspace,
    notice,
    busy,
    actions: {
      clearWorkspace,
      copySummary,
      downloadDashboard,
      downloadGraph,
      downloadWorkspace,
      importClipboard,
      importFiles,
      importPastedValue,
      loadDemo,
      printWorkspace,
      resetMembers,
      setActiveView,
      setSearch,
      shareWorkspace,
      toggleMember,
      updateSetting
    }
  };
}

function downloadJson(name: string, value: object) {
  const blob = new Blob([`${JSON.stringify(value, null, 2)}\n`], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = name;
  anchor.click();
  URL.revokeObjectURL(url);
}
