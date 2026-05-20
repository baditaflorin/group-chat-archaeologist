import type { Dashboard } from '../chat/schema';
import { dashboardSchema } from '../chat/schema';
import { workspaceStateSchema, type ImportRecord, type WorkspaceState } from './schema';

type Message = {
  id: string;
  timestamp: Date;
  sender: string;
  text: string;
  source?: string;
  line?: number;
};

type ImportWarning = {
  code: string;
  severity: string;
  message: string;
  why: string;
  nextStep: string;
  line?: number;
  evidence?: string;
};

type RawImport = {
  name: string;
  extractionMode: string;
  adapter: string;
  adapterConfidence: number;
  normalizationSteps: string[];
  warnings: ImportWarning[];
  messages: Message[];
};

type ImportOutcome =
  | { kind: 'workspace'; workspace: WorkspaceState; record: ImportRecord }
  | { kind: 'dashboard'; dashboard: Dashboard; record: ImportRecord }
  | { kind: 'raw'; raw: RawImport; record: ImportRecord }
  | { kind: 'failed'; record: ImportRecord };

const repositoryUrl = 'https://github.com/baditaflorin/group-chat-archaeologist';
const payPalUrl = 'https://www.paypal.com/paypalme/florinbadita';

const stopWords = new Set([
  'about',
  'after',
  'again',
  'also',
  'and',
  'are',
  'because',
  'been',
  'but',
  'can',
  'could',
  'did',
  'for',
  'from',
  'get',
  'had',
  'has',
  'have',
  'how',
  'into',
  'just',
  'like',
  'not',
  'now',
  'our',
  'out',
  'see',
  'she',
  'that',
  'the',
  'then',
  'there',
  'they',
  'this',
  'was',
  'we',
  'were',
  'what',
  'when',
  'with',
  'you',
  'your',
  'will'
]);

const whatsAppSenderLine =
  /^\[?(\d{1,4}[/-]\d{1,2}[/-]\d{1,4}),?\s+(\d{1,2}:\d{2}(?::\d{2})?\s?(?:[APap][Mm])?)\]?\s[-–]\s([^:]+?):\s(.*)$/;
const whatsAppHeaderLine =
  /^\[?(\d{1,4}[/-]\d{1,2}[/-]\d{1,4}),?\s+(\d{1,2}:\d{2}(?::\d{2})?\s?(?:[APap][Mm])?)\]?\s[-–]\s(.+)$/;
const iosSenderLine = /^\[(\d{1,2}\/\d{1,2}\/\d{2,4}),\s+(\d{1,2}:\d{2}(?::\d{2})?)\]\s([^:]+?):\s(.*)$/;
const isoSenderLine = /^\[?(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}(?::\d{2})?)\]?\s[-–]?\s*([^:]+?):\s(.*)$/;
const telegramMessageStart = /<div class="message[^"]*"[^>]*>/g;
const telegramFromName = /<div class="from_name">\s*(.*?)\s*<\/div>/s;
const telegramDateTitle = /<div class="date" title="([^"]+)"/s;
const telegramText = /<div class="text">\s*(.*?)\s*<\/div>/s;
const htmlTags = /<[^>]+>/g;

export async function importFromFiles(files: FileList | File[]): Promise<{
  workspace: WorkspaceState;
  records: ImportRecord[];
}> {
  const items = Array.from(files);
  const outcomes = await Promise.all(items.map((file) => importFromFile(file)));
  return buildWorkspaceFromOutcomes(outcomes);
}

export async function importFromPastedText(
  name: string,
  rawText: string,
  extractionMode: string
): Promise<{ workspace: WorkspaceState; records: ImportRecord[] }> {
  const outcome = importFromText(name, rawText, extractionMode);
  return buildWorkspaceFromOutcomes([outcome]);
}

export function importSavedStateText(name: string, rawText: string): WorkspaceState | null {
  try {
    return workspaceStateSchema.parse(JSON.parse(rawText));
  } catch {
    return null;
  }
}

export function buildSummaryText(dashboard: Dashboard): string {
  const leaders = dashboard.members
    .slice(0, 3)
    .map((member) => `${member.name} (${member.messageCount})`)
    .join(', ');
  const topTopics = dashboard.topics
    .slice(0, 3)
    .map((topic) => `${topic.label} (${topic.messageCount} messages)`)
    .join('; ');
  return [
    `Group Chat Archaeologist summary`,
    `Messages: ${dashboard.source.messageCount}`,
    `Members: ${dashboard.source.memberCount}`,
    `Adapter: ${dashboard.source.adapter ?? dashboard.source.parser}`,
    `Top members: ${leaders || 'none'}`,
    `Top topics: ${topTopics || 'none'}`
  ].join('\n');
}

async function importFromFile(file: File): Promise<ImportOutcome> {
  const text = await readFileText(file);
  return importFromText(file.name, text, extension(file.name));
}

function importFromText(name: string, rawText: string, extractionMode: string): ImportOutcome {
  const workspace = tryParseWorkspace(rawText);
  if (workspace) {
    return {
      kind: 'workspace',
      workspace,
      record: {
        name,
        kind: 'workspace',
        status: 'loaded',
        detail: 'Loaded saved browser workspace.'
      }
    };
  }

  const dashboard = tryParseDashboard(rawText);
  if (dashboard) {
    return {
      kind: 'dashboard',
      dashboard,
      record: {
        name,
        kind: 'dashboard',
        status: 'loaded',
        adapter: dashboard.source.adapter ?? dashboard.source.parser,
        messages: dashboard.source.messageCount,
        warnings: dashboard.source.warningCount ?? dashboard.warnings?.length ?? 0,
        detail: 'Loaded a dashboard JSON artifact.'
      }
    };
  }

  const normalized = normalizeText(rawText);
  const rawImport = parseRawImport(
    name,
    normalized.text,
    extractionMode,
    normalized.warnings,
    normalized.steps
  );
  if (!rawImport) {
    return {
      kind: 'failed',
      record: {
        name,
        kind: 'raw',
        status: 'failed',
        detail: 'The file did not match a supported export or saved state format.'
      }
    };
  }

  return {
    kind: 'raw',
    raw: rawImport,
    record: {
      name,
      kind: 'raw',
      status: 'loaded',
      adapter: rawImport.adapter,
      messages: rawImport.messages.length,
      warnings: rawImport.warnings.length,
      detail: 'Parsed browser-side from a chat export.'
    }
  };
}

function buildWorkspaceFromOutcomes(outcomes: ImportOutcome[]): {
  workspace: WorkspaceState;
  records: ImportRecord[];
} {
  const records = outcomes.map((outcome) => outcome.record);
  const workspaceOutcome = outcomes.find((outcome) => outcome.kind === 'workspace');
  if (workspaceOutcome && workspaceOutcome.kind === 'workspace') {
    return {
      workspace: {
        ...workspaceOutcome.workspace,
        imports: records,
        savedAt: nowIso()
      },
      records
    };
  }

  const dashboardOutcome = outcomes.find((outcome) => outcome.kind === 'dashboard');
  const raws = outcomes.filter(
    (outcome): outcome is Extract<ImportOutcome, { kind: 'raw' }> => outcome.kind === 'raw'
  );

  if (raws.length > 0) {
    const dashboard = buildDashboardFromRawImports(raws.map((item) => item.raw));
    return {
      workspace: makeWorkspace(
        'raw',
        raws.length === 1 ? raws[0].raw.name : `${raws.length} imported files`,
        dashboard,
        records
      ),
      records
    };
  }

  if (dashboardOutcome && dashboardOutcome.kind === 'dashboard') {
    return {
      workspace: makeWorkspace(
        'dashboard',
        dashboardOutcome.record.name,
        dashboardOutcome.dashboard,
        records
      ),
      records
    };
  }

  throw new Error('No importable files were loaded.');
}

function makeWorkspace(
  kind: WorkspaceState['source']['kind'],
  label: string,
  dashboard: Dashboard,
  records: ImportRecord[]
): WorkspaceState {
  return {
    schemaVersion: 'v1',
    savedAt: nowIso(),
    source: {
      kind,
      label,
      note: kind === 'demo' ? 'Shipped demo archive.' : 'Loaded from the browser.'
    },
    dashboard,
    search: '',
    selectedMembers: [],
    activeView: 'timeline',
    settings: {
      persistWorkspace: true,
      showWarnings: true,
      showConfidence: true,
      preferRestore: true
    },
    imports: records
  };
}

function tryParseWorkspace(rawText: string): WorkspaceState | null {
  try {
    return workspaceStateSchema.parse(JSON.parse(rawText));
  } catch {
    return null;
  }
}

function tryParseDashboard(rawText: string): Dashboard | null {
  try {
    return dashboardSchema.parse(JSON.parse(rawText));
  } catch {
    return null;
  }
}

async function readFileText(file: File): Promise<string> {
  const bytes = new Uint8Array(await file.arrayBuffer());
  const utf8 = new TextDecoder('utf-8', { fatal: false }).decode(bytes);
  if (utf8.includes('\uFFFD')) {
    return new TextDecoder('windows-1252').decode(bytes);
  }
  return utf8;
}

function normalizeText(input: string): { text: string; warnings: ImportWarning[]; steps: string[] } {
  let text = input;
  const steps: string[] = [];
  const warnings: ImportWarning[] = [];
  if (text.charCodeAt(0) === 0xfeff) {
    text = text.slice(1);
    warnings.push(
      makeWarning(
        'normalized_bom',
        'notice',
        'Removed a UTF-8 byte-order mark from the import.',
        'Some exports start with an invisible marker.',
        'No action needed.'
      )
    );
    steps.push('removed UTF-8 BOM');
  }
  if (text.includes('\r\n')) {
    text = text.replaceAll('\r\n', '\n');
    warnings.push(
      makeWarning(
        'normalized_crlf',
        'notice',
        'Normalized Windows line endings.',
        'Line ending differences can hide multiline message boundaries.',
        'No action needed.'
      )
    );
    steps.push('normalized CRLF');
  }
  if (text.includes('\u00A0')) {
    text = text.replaceAll('\u00A0', ' ');
    steps.push('normalized NBSP');
  }
  const cleaned = text
    .replaceAll('\u200e', '')
    .replaceAll('\u200f', '')
    .replaceAll('\u202a', '')
    .replaceAll('\u202c', '');
  if (cleaned !== text) {
    text = cleaned;
    steps.push('removed direction marks');
  }
  if (text.length > 1_000_000 || text.split('\n').length > 10_000) {
    warnings.push(
      makeWarning(
        'large_input',
        'notice',
        'Large chat export detected.',
        'This archive is big enough that browser-side analysis may take a bit longer.',
        'Let the import finish before switching tabs.'
      )
    );
  }
  if (steps.length === 0) {
    steps.push('decoded UTF-8');
  }
  return { text, warnings, steps };
}

function parseRawImport(
  name: string,
  input: string,
  extractionMode: string,
  warnings: ImportWarning[],
  normalizationSteps: string[]
): RawImport | null {
  const htmlResult = parseTelegramHtml(input);
  if (htmlResult) {
    return {
      name,
      extractionMode,
      normalizationSteps,
      ...htmlResult,
      warnings: [...warnings, ...htmlResult.warnings]
    };
  }

  const csvResult = parseDiscordCsv(input);
  if (csvResult) {
    return {
      name,
      extractionMode,
      normalizationSteps,
      ...csvResult,
      warnings: [...warnings, ...csvResult.warnings]
    };
  }

  const jsonResult = parseStructuredJson(input);
  if (jsonResult) {
    return {
      name,
      extractionMode,
      normalizationSteps,
      ...jsonResult,
      warnings: [...warnings, ...jsonResult.warnings]
    };
  }

  const textResult = parseWhatsAppText(input);
  if (textResult) {
    return {
      name,
      extractionMode,
      normalizationSteps,
      ...textResult,
      warnings: [...warnings, ...textResult.warnings]
    };
  }

  return null;
}

function parseStructuredJson(
  input: string
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  if (!input.trim().startsWith('{') && !input.trim().startsWith('[')) {
    return null;
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(input);
  } catch {
    return null;
  }

  const telegram = parseTelegramJson(parsed);
  if (telegram) {
    return telegram;
  }
  const slack = parseSlackJson(parsed);
  if (slack) {
    return slack;
  }
  const generic = parseGenericJson(parsed);
  if (generic) {
    return generic;
  }
  return null;
}

function parseTelegramJson(
  input: unknown
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  if (!isRecord(input) || !Array.isArray(input.messages)) {
    return null;
  }

  const warnings: ImportWarning[] = [];
  const messages: Message[] = [];
  for (const message of input.messages) {
    if (!isRecord(message)) {
      continue;
    }
    if (typeof message.type === 'string' && message.type !== 'message') {
      warnings.push(
        makeWarning(
          'service_message',
          'notice',
          'Skipped a Telegram service event.',
          'Service events describe group changes instead of message text.',
          'No action needed.'
        )
      );
      continue;
    }
    const sender = firstString(message.from, message.sender, message.author, message.user);
    const text = telegramTextValue(message.text);
    const timestamp = firstString(message.date, message.timestamp);
    if (!sender || !text || !timestamp) {
      continue;
    }
    const parsed = parseDate(timestamp);
    if (!parsed) {
      warnings.push(
        makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a Telegram message with an unreadable timestamp.',
          'The timestamp did not match the supported date formats.',
          'Re-export from Telegram Desktop or correct the timestamp.'
        )
      );
      continue;
    }
    messages.push(messageRecord(messages.length, parsed, sender, text, 'telegram_json'));
  }

  if (messages.length === 0) {
    return null;
  }

  return {
    adapter: 'telegram_json',
    adapterConfidence: 0.93,
    warnings,
    messages
  };
}

function parseSlackJson(
  input: unknown
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  if (!Array.isArray(input) || input.length === 0 || !isRecord(input[0])) {
    return null;
  }
  if (!('ts' in input[0]) || !('user' in input[0])) {
    return null;
  }

  const warnings: ImportWarning[] = [];
  const messages: Message[] = [];
  for (const item of input) {
    if (!isRecord(item)) {
      continue;
    }
    if (item.type && item.type !== 'message') {
      continue;
    }
    if (typeof item.subtype === 'string' && item.subtype) {
      warnings.push(
        makeWarning(
          'slack_subtype',
          'notice',
          'Skipped a Slack channel event.',
          'Subtype events are usually joins or topic changes.',
          'No action needed.'
        )
      );
      continue;
    }
    const sender = firstString(item.user);
    const text = typeof item.text === 'string' ? normalizeSlackText(item.text) : '';
    if (typeof item.text === 'string' && item.text.includes('<@')) {
      warnings.push(
        makeWarning(
          'unresolved_slack_mention',
          'notice',
          'Kept a Slack mention as a user ID.',
          'This export did not include the profile map needed for display names.',
          'Provide the Slack user export if you need names.'
        )
      );
    }
    const timestamp = typeof item.ts === 'string' ? parseSlackTimestamp(item.ts) : null;
    if (!sender || !text || !timestamp) {
      continue;
    }
    messages.push(messageRecord(messages.length, timestamp, sender, text, 'slack_json'));
  }

  if (messages.length === 0) {
    return null;
  }

  return {
    adapter: 'slack_json',
    adapterConfidence: 0.91,
    warnings,
    messages
  };
}

function parseGenericJson(
  input: unknown
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  const rows = Array.isArray(input)
    ? input
    : isRecord(input) && Array.isArray(input.messages)
      ? input.messages
      : null;
  if (!rows) {
    return null;
  }

  const warnings: ImportWarning[] = [];
  const messages: Message[] = [];
  for (const item of rows) {
    if (!isRecord(item)) {
      continue;
    }
    const sender = firstString(item.sender, item.author, item.from, item.user);
    const text = firstString(item.text, item.message, item.content);
    const timestamp = firstString(item.timestamp, item.date);
    if (!sender || !text || !timestamp) {
      continue;
    }
    const parsed = parseDate(timestamp);
    if (!parsed) {
      warnings.push(
        makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a JSON message with an unreadable timestamp.',
          'The timestamp did not match the supported date formats.',
          'Correct the timestamp or export with ISO-8601 dates.'
        )
      );
      continue;
    }
    messages.push(messageRecord(messages.length, parsed, sender, text, 'json'));
  }

  if (messages.length === 0) {
    return null;
  }

  return {
    adapter: 'json',
    adapterConfidence: 0.74,
    warnings,
    messages
  };
}

function parseDiscordCsv(
  input: string
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  const lines = input.trim().split('\n');
  if (lines.length < 2) {
    return null;
  }
  const header = splitCsvLine(lines[0]).map((item) => item.trim().toLowerCase());
  const dateIndex = header.indexOf('date');
  const authorIndex = header.findIndex(
    (item) => item === 'author' || item === 'authorid' || item === 'sender'
  );
  const contentIndex = header.findIndex(
    (item) => item === 'content' || item === 'message' || item === 'text'
  );
  if (dateIndex === -1 || authorIndex === -1 || contentIndex === -1) {
    return null;
  }

  const warnings: ImportWarning[] = [];
  const messages: Message[] = [];
  for (let index = 1; index < lines.length; index += 1) {
    const row = splitCsvLine(lines[index]);
    if (row.length <= Math.max(dateIndex, authorIndex, contentIndex)) {
      continue;
    }
    const text = row[contentIndex].trim();
    if (text.includes('\n')) {
      warnings.push(
        makeWarning(
          'csv_multiline_field',
          'notice',
          'Preserved a multiline Discord message.',
          'CSV quoting allows line breaks inside a message body.',
          'No action needed.',
          index + 1
        )
      );
    }
    const parsed = parseDate(row[dateIndex].trim());
    if (!parsed) {
      warnings.push(
        makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a Discord row with an unreadable timestamp.',
          'The Date column did not match the supported date formats.',
          'Export with ISO-8601 dates or correct the row.',
          index + 1
        )
      );
      continue;
    }
    const sender = row[authorIndex].trim();
    if (!sender || !text) {
      continue;
    }
    messages.push(messageRecord(messages.length, parsed, sender, text, 'discord_csv', index + 1));
  }

  if (messages.length === 0) {
    return null;
  }

  return {
    adapter: 'discord_csv',
    adapterConfidence: 0.9,
    warnings,
    messages
  };
}

function parseTelegramHtml(
  input: string
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  if (!input.toLowerCase().includes('from_name') || !input.toLowerCase().includes('class="message')) {
    return null;
  }

  const warnings: ImportWarning[] = [];
  const starts = [...input.matchAll(telegramMessageStart)];
  const messages: Message[] = [];
  for (let index = 0; index < starts.length; index += 1) {
    const start = starts[index].index ?? 0;
    const next = starts[index + 1]?.index ?? input.length;
    const chunk = input.slice(start, next);
    const sender = cleanHtmlMatch(chunk.match(telegramFromName)?.[1]);
    const stamp = cleanHtmlMatch(chunk.match(telegramDateTitle)?.[1]);
    const text = cleanHtmlMatch(chunk.match(telegramText)?.[1]);
    if (!sender || !stamp || !text) {
      continue;
    }
    const parsed = parseDate(stamp);
    if (!parsed) {
      warnings.push(
        makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a Telegram HTML message with an unreadable timestamp.',
          'The HTML date title did not match the supported date formats.',
          'Re-export from Telegram Desktop or correct the date.',
          index + 1
        )
      );
      continue;
    }
    messages.push(messageRecord(messages.length, parsed, sender, text, 'telegram_html'));
  }

  if (messages.length === 0) {
    return null;
  }

  return {
    adapter: 'telegram_html',
    adapterConfidence: 0.86,
    warnings,
    messages
  };
}

function parseWhatsAppText(
  input: string
): Omit<RawImport, 'name' | 'extractionMode' | 'normalizationSteps'> | null {
  const warnings: ImportWarning[] = [];
  const messages: Message[] = [];
  const lines = input.split('\n');
  let matchedHeaders = 0;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index].trimEnd();
    if (!line.trim()) {
      continue;
    }
    const parsed = parseWhatsAppLine(line, messages.length, index + 1);
    if (parsed) {
      matchedHeaders += 1;
      if (parsed.warning) {
        warnings.push(parsed.warning);
      }
      if (parsed.message) {
        messages.push(parsed.message);
      }
      continue;
    }

    if (messages.length === 0) {
      warnings.push(
        makeWarning(
          'orphan_line',
          'warning',
          'Ignored text before the first recognized message.',
          'There was no earlier message to attach this line to.',
          'Check whether the export starts in the middle of a message.',
          index + 1
        )
      );
      continue;
    }

    messages[messages.length - 1].text += `\n${line.trim()}`;
  }

  if (messages.length === 0 && matchedHeaders === 0) {
    return null;
  }

  return {
    adapter: 'whatsapp_text',
    adapterConfidence: Math.min(0.97, 0.65 + matchedHeaders / 20),
    warnings,
    messages
  };
}

function parseWhatsAppLine(
  line: string,
  nextIndex: number,
  lineNumber: number
): { message?: Message; warning?: ImportWarning } | null {
  const isoMatch = line.match(isoSenderLine);
  if (isoMatch) {
    const parsed = parseDate(isoMatch[1]);
    if (!parsed) {
      return {
        warning: makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a message with an unreadable timestamp.',
          'The line looked like a message, but the timestamp could not be parsed.',
          'Check the timestamp format.',
          lineNumber,
          line
        )
      };
    }
    return {
      message: messageRecord(nextIndex, parsed, isoMatch[2], isoMatch[3], 'whatsapp_text', lineNumber)
    };
  }

  const iosMatch = line.match(iosSenderLine);
  if (iosMatch) {
    const parsed = parseDate(`${iosMatch[1]} ${iosMatch[2]}`);
    if (!parsed) {
      return {
        warning: makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a message with an unreadable timestamp.',
          'The line looked like an iOS export entry, but the timestamp could not be parsed.',
          'Check the timestamp format.',
          lineNumber,
          line
        )
      };
    }
    return {
      message: messageRecord(nextIndex, parsed, iosMatch[3], iosMatch[4], 'whatsapp_text', lineNumber)
    };
  }

  const whatsAppMatch = line.match(whatsAppSenderLine);
  if (whatsAppMatch) {
    const parsed = parseDate(`${whatsAppMatch[1]} ${whatsAppMatch[2]}`);
    if (!parsed) {
      return {
        warning: makeWarning(
          'unparsed_timestamp',
          'warning',
          'Skipped a message with an unreadable timestamp.',
          'The line looked like a message, but the timestamp could not be parsed.',
          'Check the timestamp format.',
          lineNumber,
          line
        )
      };
    }
    const message = messageRecord(
      nextIndex,
      parsed,
      whatsAppMatch[3],
      whatsAppMatch[4],
      'whatsapp_text',
      lineNumber
    );
    if (message.text.toLowerCase().includes('media omitted')) {
      return {
        message,
        warning: makeWarning(
          'media_placeholder',
          'notice',
          'Found a media placeholder.',
          'The export references an attachment but does not include the media file.',
          'Re-export with media if the attachment matters.',
          lineNumber,
          message.text
        )
      };
    }
    return { message };
  }

  const headerMatch = line.match(whatsAppHeaderLine);
  if (headerMatch) {
    const detail = headerMatch[3].trim();
    if (looksLikeMalformedSender(detail)) {
      return {
        warning: makeWarning(
          'malformed_timestamp_line',
          'warning',
          'Skipped a timestamp line with no message body.',
          'The line has a timestamp and likely a sender, but no message text.',
          'Check whether the export was truncated around this line.',
          lineNumber,
          line
        )
      };
    }
    return {
      warning: makeWarning(
        'system_message',
        'notice',
        'Skipped a WhatsApp system message.',
        'System messages describe encryption, group changes, or settings instead of conversation text.',
        'No action needed unless you want system events included.',
        lineNumber,
        line
      )
    };
  }

  return null;
}

function buildDashboardFromRawImports(rawImports: RawImport[]): Dashboard {
  const messages = rawImports
    .flatMap((item) => item.messages)
    .sort((left, right) => left.timestamp.getTime() - right.timestamp.getTime())
    .map((message, index) => ({ ...message, id: `m_${String(index + 1).padStart(6, '0')}` }));

  const warnings = rawImports.flatMap((item) => item.warnings);
  const memberStats = buildMemberStats(messages);
  const topics = buildTopics(messages);
  const introductions = buildIntroductions(messages);
  const insideJokes = buildInsideJokes(messages);
  const departures = buildDepartures(messages);
  const notableMessages = buildNotableMessages(messages);
  const graph = buildGraph(
    messages,
    memberStats.map((member) => member.name),
    introductions
  );
  const adapters = Array.from(new Set(rawImports.map((item) => item.adapter)));
  const label = rawImports.length === 1 ? rawImports[0].name : `${rawImports.length} imported files`;
  const firstAt = messages[0]?.timestamp.toISOString() ?? '';
  const lastAt = messages[messages.length - 1]?.timestamp.toISOString() ?? '';

  return {
    schemaVersion: 'v1',
    generatedAt: nowIso(),
    repositoryUrl,
    paypalUrl: payPalUrl,
    source: {
      inputName: label,
      inputSha256: simpleHash(
        messages
          .map((message) => `${message.sender}|${message.timestamp.toISOString()}|${message.text}`)
          .join('\n')
      ),
      parser: adapters.join(','),
      adapter: adapters.join(','),
      adapterConfidence: round(
        rawImports.reduce((sum, item) => sum + item.adapterConfidence, 0) / rawImports.length
      ),
      extractionMode: Array.from(new Set(rawImports.map((item) => item.extractionMode))).join(','),
      normalizationSteps: Array.from(new Set(rawImports.flatMap((item) => item.normalizationSteps))),
      analyticsEngine: 'browser-workspace',
      messageCount: messages.length,
      memberCount: memberStats.length,
      firstMessageAt: firstAt,
      lastMessageAt: lastAt,
      warningCount: warnings.length,
      llmProvider: 'heuristic',
      llmModel: 'none',
      llmUsed: false,
      sourceCommit: 'browser',
      appVersion: __APP_VERSION__,
      parameters: {
        mode: 'browser-import',
        importedFiles: String(rawImports.length)
      }
    },
    members: memberStats.map((member) => ({
      name: member.name,
      messageCount: member.count,
      firstMessageAt: member.firstAt.toISOString(),
      lastMessageAt: member.lastAt.toISOString()
    })),
    topics,
    introductions,
    insideJokes,
    departures,
    notableMessages,
    warnings,
    debug: {
      adapterEvidence: rawImports.map((item) => `${item.name}: ${item.adapter}`),
      parseWarnings: warnings.length
    },
    graph
  };
}

function buildMemberStats(messages: Message[]) {
  const members = new Map<string, { name: string; count: number; firstAt: Date; lastAt: Date }>();
  for (const message of messages) {
    const current = members.get(message.sender);
    if (!current) {
      members.set(message.sender, {
        name: message.sender,
        count: 1,
        firstAt: message.timestamp,
        lastAt: message.timestamp
      });
      continue;
    }
    current.count += 1;
    current.lastAt = message.timestamp;
  }
  return [...members.values()].sort(
    (left, right) => right.count - left.count || left.name.localeCompare(right.name)
  );
}

function buildTopics(messages: Message[]): Dashboard['topics'] {
  const buckets = new Map<
    string,
    { start: Date; end: Date; members: Map<string, number>; words: Map<string, number>; count: number }
  >();
  for (const message of messages) {
    const key = `${message.timestamp.getUTCFullYear()}-${String(message.timestamp.getUTCMonth() + 1).padStart(2, '0')}`;
    const start = new Date(Date.UTC(message.timestamp.getUTCFullYear(), message.timestamp.getUTCMonth(), 1));
    const end = new Date(
      Date.UTC(message.timestamp.getUTCFullYear(), message.timestamp.getUTCMonth() + 1, 0)
    );
    const bucket = buckets.get(key) ?? { start, end, members: new Map(), words: new Map(), count: 0 };
    bucket.count += 1;
    bucket.members.set(message.sender, (bucket.members.get(message.sender) ?? 0) + 1);
    for (const token of tokenize(message.text)) {
      if (!stopWords.has(token)) {
        bucket.words.set(token, (bucket.words.get(token) ?? 0) + 1);
      }
    }
    buckets.set(key, bucket);
  }

  return [...buckets.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, bucket]) => {
      const keywords = topKeys(bucket.words, 5);
      const topMembers = topKeys(bucket.members, 3);
      return {
        id: `topic_${key.replace('-', '_')}`,
        label: keywords.length > 0 ? titleFromKeywords(keywords) : 'General memory',
        start: bucket.start.toISOString().slice(0, 10),
        end: bucket.end.toISOString().slice(0, 10),
        messageCount: bucket.count,
        keywords,
        topMembers,
        summary: `${bucket.count} messages led by ${topMembers.join(', ') || 'nobody in particular'}.`,
        confidence: {
          score: round(Math.min(0.9, 0.55 + bucket.count / 40)),
          level: confidenceLevel(Math.min(0.9, 0.55 + bucket.count / 40)),
          evidence: [
            `month bucket ${key} contains ${bucket.count} parsed messages`,
            `top keywords: ${keywords.join(', ')}`
          ]
        }
      };
    });
}

function buildIntroductions(messages: Message[]): Dashboard['introductions'] {
  const firstSent = new Map<string, Date>();
  for (const message of messages) {
    if (!firstSent.has(message.sender)) {
      firstSent.set(message.sender, message.timestamp);
    }
  }

  const members = [...firstSent.keys()].sort((left, right) => left.localeCompare(right));
  const seen = new Set<string>();
  const edges: Dashboard['introductions'] = [];

  for (const message of messages) {
    const body = message.text.toLowerCase();
    for (const target of members) {
      const targetFirst = firstSent.get(target);
      if (!targetFirst || target === message.sender || seen.has(target) || message.timestamp >= targetFirst) {
        continue;
      }
      if (new RegExp(`\\b${escapeRegExp(target.toLowerCase())}\\b`).test(body)) {
        seen.add(target);
        edges.push({
          from: message.sender,
          to: target,
          firstMentionAt: message.timestamp.toISOString(),
          messageId: message.id,
          snippet: snippet(message.text, 140),
          confidence: {
            score: 0.78,
            level: 'medium',
            evidence: [
              `${target} was mentioned before their first message`,
              `first ${target} message was at ${targetFirst.toISOString()}`
            ]
          }
        });
      }
    }
  }

  return edges;
}

function buildInsideJokes(messages: Message[]): Dashboard['insideJokes'] {
  const stats = new Map<string, { count: number; origin: Message; participants: Set<string> }>();
  for (const message of messages) {
    const tokens = tokenize(message.text);
    const seen = new Set<string>();
    for (let size = 2; size <= 4; size += 1) {
      for (let index = 0; index + size <= tokens.length; index += 1) {
        const phraseTokens = tokens.slice(index, index + size);
        if (phraseTokens.every((token) => stopWords.has(token) || token.length <= 2)) {
          continue;
        }
        const phrase = phraseTokens.join(' ');
        if (seen.has(phrase)) {
          continue;
        }
        seen.add(phrase);
        const current = stats.get(phrase) ?? { count: 0, origin: message, participants: new Set<string>() };
        current.count += 1;
        current.participants.add(message.sender);
        stats.set(phrase, current);
      }
    }
  }

  return [...stats.entries()]
    .filter(([, value]) => value.count >= 2 && value.participants.size >= 2)
    .sort(
      (left, right) =>
        right[1].count - left[1].count ||
        left[1].origin.timestamp.getTime() - right[1].origin.timestamp.getTime()
    )
    .filter(
      (entry, index, entries) =>
        !entries
          .slice(0, index)
          .some(([existing, value]) => value.count === entry[1].count && existing.includes(entry[0]))
    )
    .slice(0, 8)
    .map(([phrase, value]) => ({
      phrase,
      originAt: value.origin.timestamp.toISOString(),
      originSender: value.origin.sender,
      originId: value.origin.id,
      occurrences: value.count,
      participants: [...value.participants].sort((left, right) => left.localeCompare(right)),
      snippet: snippet(value.origin.text, 160),
      confidence: {
        score: round(Math.min(0.9, 0.45 + (value.count + value.participants.size) / 12)),
        level: confidenceLevel(Math.min(0.9, 0.45 + (value.count + value.participants.size) / 12)),
        evidence: [`phrase repeated ${value.count} times`, `used by ${value.participants.size} participants`]
      }
    }));
}

function buildDepartures(messages: Message[]): Dashboard['departures'] {
  if (messages.length === 0) {
    return [];
  }
  const byMember = new Map<string, { first: Message; last: Message; count: number }>();
  for (const message of messages) {
    const current = byMember.get(message.sender);
    if (!current) {
      byMember.set(message.sender, { first: message, last: message, count: 1 });
      continue;
    }
    current.last = message;
    current.count += 1;
  }
  const finalAt = messages[messages.length - 1].timestamp;
  return [...byMember.entries()]
    .map(([member, value]) => {
      const daysQuiet = Math.max(
        0,
        Math.round((finalAt.getTime() - value.last.timestamp.getTime()) / 86_400_000)
      );
      const activeSpanDays = Math.max(
        0,
        Math.round((value.last.timestamp.getTime() - value.first.timestamp.getTime()) / 86_400_000)
      );
      const status = daysQuiet > 180 ? 'departed' : daysQuiet > 60 ? 'quiet' : 'active';
      const interpretation =
        status === 'departed'
          ? 'No recent activity in the observed archive window.'
          : status === 'quiet'
            ? 'Activity faded before the archive ended.'
            : 'Still active near the end of the archive.';
      const score = status === 'active' ? 0.64 : status === 'quiet' ? 0.72 : 0.84;
      return {
        member,
        status,
        lastMessageAt: value.last.timestamp.toISOString(),
        daysSinceActive: daysQuiet,
        activeSpanDays,
        lastSnippet: snippet(value.last.text, 140),
        interpretation,
        confidence: {
          score,
          level: confidenceLevel(score),
          evidence: [
            `last message is ${daysQuiet} days before archive end`,
            `${value.count} messages across ${activeSpanDays} active days`
          ]
        }
      };
    })
    .sort(
      (left, right) => right.daysSinceActive - left.daysSinceActive || left.member.localeCompare(right.member)
    );
}

function buildNotableMessages(messages: Message[]): Dashboard['notableMessages'] {
  const firstByMember = new Set<string>();
  const notable: Dashboard['notableMessages'] = [];
  for (const message of messages) {
    if (!firstByMember.has(message.sender)) {
      firstByMember.add(message.sender);
      notable.push({
        id: message.id,
        at: message.timestamp.toISOString(),
        sender: message.sender,
        kind: 'first-message',
        snippet: snippet(message.text, 150),
        why: 'First message by this member in the archive.'
      });
    }
  }
  const longest = messages.reduce<Message | null>((current, message) => {
    if (!current || message.text.length > current.text.length) {
      return message;
    }
    return current;
  }, null);
  if (longest) {
    notable.push({
      id: longest.id,
      at: longest.timestamp.toISOString(),
      sender: longest.sender,
      kind: 'longest-message',
      snippet: snippet(longest.text, 180),
      why: 'Longest message in the archive.'
    });
  }
  return notable.slice(0, 10);
}

function buildGraph(
  messages: Message[],
  members: string[],
  introductions: Dashboard['introductions']
): Dashboard['graph'] {
  const nodes = members.sort((left, right) => left.localeCompare(right));
  const radius = 180;
  const center = 220;
  const points = nodes.map((name, index) => {
    const angle = (Math.PI * 2 * index) / Math.max(1, nodes.length);
    return {
      name,
      x: center + Math.cos(angle - Math.PI / 2) * radius,
      y: center + Math.sin(angle - Math.PI / 2) * radius
    };
  });
  const pointByName = new Map(points.map((point) => [point.name, point]));
  const lines = introductions
    .map((edge) => {
      const from = pointByName.get(edge.from);
      const to = pointByName.get(edge.to);
      if (!from || !to) {
        return '';
      }
      return `<line x1="${from.x}" y1="${from.y}" x2="${to.x}" y2="${to.y}" stroke="#2f6f89" stroke-width="2" stroke-linecap="round" />`;
    })
    .join('');
  const circles = points
    .map(
      (point) =>
        `<g><circle cx="${point.x}" cy="${point.y}" r="22" fill="#2f5f53" /><text x="${point.x}" y="${point.y + 5}" font-size="10" text-anchor="middle" fill="#ffffff">${escapeXml(point.name.slice(0, 10))}</text></g>`
    )
    .join('');
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="440" height="440" viewBox="0 0 440 440"><rect width="440" height="440" fill="#fbfaf6" rx="16" /><text x="220" y="28" font-size="16" text-anchor="middle" fill="#16251f">Who Introduced Whom</text>${lines}${circles}</svg>`;
  return {
    dotPath: '',
    svgPath: `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`,
    rendered: true,
    renderer: messages.length > 0 ? 'browser-svg' : 'none'
  };
}

function messageRecord(
  index: number,
  timestamp: Date,
  sender: string,
  text: string,
  source: string,
  line?: number
): Message {
  return {
    id: `m_${String(index + 1).padStart(6, '0')}`,
    timestamp,
    sender: sender.trim(),
    text: text.trim(),
    source,
    line
  };
}

function splitCsvLine(input: string): string[] {
  const out: string[] = [];
  let current = '';
  let inQuotes = false;
  for (let index = 0; index < input.length; index += 1) {
    const char = input[index];
    if (char === '"' && input[index + 1] === '"') {
      current += '"';
      index += 1;
      continue;
    }
    if (char === '"') {
      inQuotes = !inQuotes;
      continue;
    }
    if (char === ',' && !inQuotes) {
      out.push(current);
      current = '';
      continue;
    }
    current += char;
  }
  out.push(current);
  return out;
}

function tokenize(text: string) {
  return (
    text
      .toLowerCase()
      .replaceAll("'", ' ')
      .replaceAll('"', ' ')
      .replaceAll('’', ' ')
      .replaceAll('-', ' ')
      .match(/[a-z0-9]+/g) ?? []
  );
}

function topKeys(map: Map<string, number>, limit: number) {
  return [...map.entries()]
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .slice(0, limit)
    .map(([key]) => key);
}

function titleFromKeywords(keywords: string[]) {
  return keywords
    .slice(0, 3)
    .map((keyword) => keyword.charAt(0).toUpperCase() + keyword.slice(1))
    .join(' / ');
}

function snippet(text: string, limit: number) {
  const compact = text.trim().replace(/\s+/g, ' ');
  if (compact.length <= limit) {
    return compact;
  }
  return `${compact.slice(0, limit - 1).trim()}...`;
}

function cleanHtmlMatch(value: string | undefined) {
  if (!value) {
    return '';
  }
  return value
    .replace(htmlTags, ' ')
    .replaceAll('&amp;', '&')
    .replaceAll('&quot;', '"')
    .replaceAll('&#39;', "'")
    .replaceAll('&lt;', '<')
    .replaceAll('&gt;', '>')
    .replace(/\s+/g, ' ')
    .trim();
}

function normalizeSlackText(value: string) {
  return value
    .replace(/<@([^>]+)>/g, '$1')
    .replaceAll('&amp;', '&')
    .trim();
}

function telegramTextValue(value: unknown): string {
  if (typeof value === 'string') {
    return value;
  }
  if (!Array.isArray(value)) {
    return '';
  }
  return value
    .map((item) => {
      if (typeof item === 'string') {
        return item;
      }
      if (isRecord(item) && typeof item.text === 'string') {
        return item.text;
      }
      return '';
    })
    .join('');
}

function parseDate(value: string): Date | null {
  const cleaned = value
    .replace(/ UTC[+-]\d{2}:\d{2}$/, '')
    .replace(/\s+/g, ' ')
    .trim();
  const native = new Date(cleaned);
  if (!Number.isNaN(native.getTime())) {
    return native;
  }

  const layouts = [
    /^(\d{1,2})\/(\d{1,2})\/(\d{2,4}) (\d{1,2}):(\d{2})(?::(\d{2}))? ?([AP]M)?$/i,
    /^(\d{1,2})-(\d{1,2})-(\d{2,4}) (\d{1,2}):(\d{2})(?::(\d{2}))?$/,
    /^(\d{2})\.(\d{2})\.(\d{4}) (\d{2}):(\d{2})(?::(\d{2}))?$/
  ];

  for (const layout of layouts) {
    const match = cleaned.match(layout);
    if (!match) {
      continue;
    }
    if (layout === layouts[0]) {
      let month = Number(match[1]);
      let day = Number(match[2]);
      const year = normalizeYear(Number(match[3]));
      let hour = Number(match[4]);
      const minute = Number(match[5]);
      const second = Number(match[6] ?? '0');
      const suffix = match[7]?.toUpperCase();
      if (suffix === 'PM' && hour < 12) {
        hour += 12;
      } else if (suffix === 'AM' && hour === 12) {
        hour = 0;
      }
      if (month > 12) {
        [month, day] = [day, month];
      }
      return new Date(Date.UTC(year, month - 1, day, hour, minute, second));
    }
    if (layout === layouts[1]) {
      const day = Number(match[1]);
      const month = Number(match[2]);
      const year = normalizeYear(Number(match[3]));
      return new Date(
        Date.UTC(year, month - 1, day, Number(match[4]), Number(match[5]), Number(match[6] ?? '0'))
      );
    }
    const day = Number(match[1]);
    const month = Number(match[2]);
    const year = Number(match[3]);
    return new Date(
      Date.UTC(year, month - 1, day, Number(match[4]), Number(match[5]), Number(match[6] ?? '0'))
    );
  }
  return null;
}

function parseSlackTimestamp(value: string): Date | null {
  const parsed = Number(value);
  if (Number.isNaN(parsed)) {
    return null;
  }
  return new Date(parsed * 1000);
}

function normalizeYear(value: number) {
  return value < 100 ? 2000 + value : value;
}

function looksLikeMalformedSender(value: string) {
  return !value.includes(':') && value.trim().split(/\s+/).length <= 3;
}

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value.trim();
    }
  }
  return '';
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function makeWarning(
  code: string,
  severity: string,
  message: string,
  why: string,
  nextStep: string,
  line?: number,
  evidence?: string
): ImportWarning {
  return { code, severity, message, why, nextStep, line, evidence };
}

function simpleHash(value: string) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(index);
    hash |= 0;
  }
  return Math.abs(hash).toString(16).padStart(8, '0');
}

function round(value: number) {
  return Math.round(value * 100) / 100;
}

function confidenceLevel(score: number): 'low' | 'medium' | 'high' {
  if (score >= 0.8) {
    return 'high';
  }
  if (score >= 0.6) {
    return 'medium';
  }
  return 'low';
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function escapeXml(value: string) {
  return value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');
}

function extension(name: string) {
  const index = name.lastIndexOf('.');
  return index >= 0 ? name.slice(index + 1).toLowerCase() : 'txt';
}

function nowIso() {
  return new Date().toISOString();
}
