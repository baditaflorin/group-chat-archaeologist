import {
  AlertTriangle,
  CalendarClock,
  CircleHelp,
  GitFork,
  Github,
  HeartHandshake,
  MessageSquareQuote,
  Network,
  Search,
  Sparkles,
  UsersRound
} from 'lucide-react';
import { useState } from 'react';
import { assetUrl } from './features/chat/api';
import { daysAgo, formatDate, formatMonth } from './features/chat/format';
import type { Dashboard, Departure, InsideJoke, Introduction, TopicPeriod } from './features/chat/schema';
import { useBuildInfo } from './features/chat/useBuildInfo';
import { useDashboard, useMeta } from './features/chat/useDashboard';

type View = 'timeline' | 'map' | 'jokes' | 'departures';

const views: Array<{ id: View; label: string; icon: typeof CalendarClock }> = [
  { id: 'timeline', label: 'Timeline', icon: CalendarClock },
  { id: 'map', label: 'Map', icon: Network },
  { id: 'jokes', label: 'Origins', icon: MessageSquareQuote },
  { id: 'departures', label: 'Departures', icon: UsersRound }
];

export function App() {
  const dashboard = useDashboard();
  const meta = useMeta();
  const buildInfo = useBuildInfo();
  const [view, setView] = useState<View>(() => (localStorage.getItem('active-view') as View) || 'timeline');
  const [search, setSearch] = useState('');
  const [members, setMembers] = useState<string[]>([]);

  function selectView(next: View) {
    setView(next);
    localStorage.setItem('active-view', next);
  }

  if (dashboard.isLoading) {
    return <ShellState title="Reading the archive map" body="Loading static artifacts from GitHub Pages." />;
  }

  if (dashboard.isError || !dashboard.data) {
    return (
      <ShellState
        title="The archive map could not load"
        body={dashboard.error instanceof Error ? dashboard.error.message : 'Unknown data loading error.'}
      />
    );
  }

  const data = dashboard.data;
  const warnings = data.warnings ?? [];
  const debugMode = new URLSearchParams(window.location.search).has('debug');
  const activeMembers = members.length === 0 ? data.members.map((member) => member.name) : members;
  const filtered = filterData(data, search, activeMembers);

  return (
    <main className="min-h-screen bg-paper text-ink">
      <Header data={data} updatedAt={meta.data?.generatedAt ?? data.generatedAt} />
      <section className="mx-auto grid w-full max-w-7xl gap-5 px-4 pb-8 pt-4 lg:grid-cols-[280px_minmax(0,1fr)]">
        <aside className="self-start rounded-md border border-ink/10 bg-white/80 p-4 shadow-soft">
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3 top-3 h-4 w-4 text-ink/50"
              aria-hidden="true"
            />
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="h-10 w-full rounded-md border border-ink/15 bg-white pl-9 pr-3 text-sm outline-none ring-moss/30 focus:ring-4"
              placeholder="Search topics, jokes, people"
              aria-label="Search archive"
            />
          </div>

          <div className="mt-5">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-ink/60">Members</h2>
            <div className="mt-3 grid gap-2">
              {data.members.map((member) => {
                const checked = activeMembers.includes(member.name);
                return (
                  <label
                    key={member.name}
                    className="flex items-center justify-between gap-3 rounded-md px-2 py-2 hover:bg-moss/10"
                  >
                    <span className="flex items-center gap-2 text-sm font-medium">
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => {
                          setMembers((current) =>
                            checked
                              ? data.members.map((item) => item.name).filter((name) => name !== member.name)
                              : Array.from(new Set([...current, member.name]))
                          );
                        }}
                        className="h-4 w-4 rounded border-ink/20 accent-moss"
                      />
                      {member.name}
                    </span>
                    <span className="text-xs tabular-nums text-ink/55">{member.messageCount}</span>
                  </label>
                );
              })}
            </div>
            <button
              className="mt-3 inline-flex h-9 items-center gap-2 rounded-md border border-ink/10 px-3 text-sm font-semibold text-moss hover:bg-moss/10"
              onClick={() => setMembers([])}
            >
              <Sparkles className="h-4 w-4" aria-hidden="true" />
              Reset
            </button>
          </div>

          <SourcePanel data={data} />
        </aside>

        <section className="min-w-0">
          <nav
            className="grid grid-cols-2 gap-2 rounded-md border border-ink/10 bg-white/70 p-2 md:grid-cols-4"
            aria-label="Archive views"
          >
            {views.map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.id}
                  onClick={() => selectView(item.id)}
                  className={`flex h-11 items-center justify-center gap-2 rounded-md text-sm font-semibold transition ${
                    view === item.id ? 'bg-ink text-white' : 'text-ink/70 hover:bg-ink/10'
                  }`}
                >
                  <Icon className="h-4 w-4" aria-hidden="true" />
                  {item.label}
                </button>
              );
            })}
          </nav>

          <div className="mt-5">
            {warnings.length > 0 && <WarningPanel warnings={warnings} />}
            {debugMode && <DebugPanel data={data} />}
            {view === 'timeline' && <Timeline topics={filtered.topics} />}
            {view === 'map' && <MapView data={data} introductions={filtered.introductions} />}
            {view === 'jokes' && <Jokes jokes={filtered.insideJokes} />}
            {view === 'departures' && <Departures departures={filtered.departures} />}
          </div>
        </section>
      </section>
      <Footer data={data} commit={buildInfo.data?.commit ?? data.source.sourceCommit} />
    </main>
  );
}

function Header({ data, updatedAt }: { data: Dashboard; updatedAt: string }) {
  return (
    <header className="border-b border-ink/10 bg-[#fbfaf6]">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-5 px-4 py-5 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="text-sm font-semibold uppercase tracking-wide text-rust">Shared memory, surfaced</p>
          <h1 className="mt-1 text-4xl font-black leading-tight text-ink md:text-6xl">
            Group Chat Archaeologist
          </h1>
          <p className="mt-3 max-w-3xl text-base leading-7 text-ink/70">
            Topic timeline, who-introduced-whom map, inside-joke origin tracer, and member-departure analysis
            from a static demo archive.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <a
            className="inline-flex h-10 items-center gap-2 rounded-md bg-ink px-4 text-sm font-bold text-white hover:bg-moss"
            href={data.repositoryUrl}
            target="_blank"
            rel="noreferrer"
          >
            <Github className="h-4 w-4" aria-hidden="true" />
            Star on GitHub
          </a>
          <a
            className="inline-flex h-10 items-center gap-2 rounded-md border border-rust/30 bg-white px-4 text-sm font-bold text-rust hover:bg-rust hover:text-white"
            href={data.paypalUrl}
            target="_blank"
            rel="noreferrer"
          >
            <HeartHandshake className="h-4 w-4" aria-hidden="true" />
            Support
          </a>
        </div>
      </div>
      <div className="mx-auto grid w-full max-w-7xl gap-3 px-4 pb-5 md:grid-cols-5">
        <Metric label="Messages" value={data.source.messageCount.toLocaleString()} />
        <Metric label="Members" value={data.source.memberCount.toLocaleString()} />
        <Metric label="Topics" value={data.topics.length.toLocaleString()} />
        <Metric
          label="Warnings"
          value={(data.source.warningCount ?? data.warnings?.length ?? 0).toLocaleString()}
        />
        <Metric label="Updated" value={daysAgo(updatedAt)} />
      </div>
    </header>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-ink/10 bg-white px-4 py-3">
      <div className="text-xs font-semibold uppercase tracking-wide text-ink/50">{label}</div>
      <div className="mt-1 text-2xl font-black text-ink">{value}</div>
    </div>
  );
}

function SourcePanel({ data }: { data: Dashboard }) {
  return (
    <div className="mt-6 border-t border-ink/10 pt-5 text-sm text-ink/70">
      <h2 className="font-semibold uppercase tracking-wide text-ink/60">Artifact</h2>
      <dl className="mt-3 grid gap-2">
        <Info label="Schema" value={data.schemaVersion} />
        <Info
          label="Adapter"
          value={`${data.source.adapter ?? data.source.parser} (${percent(data.source.adapterConfidence)})`}
        />
        <Info label="Extract" value={data.source.extractionMode} />
        <Info label="Analytics" value={data.source.analyticsEngine} />
        <Info label="LLM" value={data.source.llmUsed ? data.source.llmModel : 'heuristic fallback'} />
        <Info label="Warnings" value={`${data.source.warningCount ?? data.warnings?.length ?? 0}`} />
        <Info label="Input SHA" value={data.source.inputSha256.slice(0, 10)} />
      </dl>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <dt className="text-ink/50">{label}</dt>
      <dd className="max-w-[150px] truncate font-mono text-xs text-ink">{value}</dd>
    </div>
  );
}

function WarningPanel({ warnings }: { warnings: NonNullable<Dashboard['warnings']> }) {
  const topWarnings = warnings.slice(0, 4);
  return (
    <section className="mb-5 rounded-md border border-rust/25 bg-rust/5 p-4">
      <h2 className="flex items-center gap-2 text-sm font-black uppercase tracking-wide text-rust">
        <AlertTriangle className="h-4 w-4" aria-hidden="true" />
        Import Warnings
      </h2>
      <div className="mt-3 grid gap-3">
        {topWarnings.map((warning, index) => (
          <article key={`${warning.code}-${warning.line ?? index}`} className="text-sm leading-6 text-ink/75">
            <div className="font-bold text-ink">
              {warning.message}
              {warning.line ? (
                <span className="font-mono text-xs text-ink/45"> line {warning.line}</span>
              ) : null}
            </div>
            <div>{warning.why}</div>
            <div className="font-semibold text-rust">{warning.nextStep}</div>
          </article>
        ))}
      </div>
    </section>
  );
}

function DebugPanel({ data }: { data: Dashboard }) {
  return (
    <section className="mb-5 rounded-md border border-lake/25 bg-lake/5 p-4 text-sm text-ink/75">
      <h2 className="font-black uppercase tracking-wide text-lake">Debug</h2>
      <dl className="mt-3 grid gap-2 md:grid-cols-2">
        <Info label="Adapter evidence" value={(data.debug?.adapterEvidence ?? []).join('; ') || 'none'} />
        <Info label="Normalize" value={(data.source.normalizationSteps ?? []).join('; ') || 'none'} />
        <Info
          label="Parameters"
          value={
            Object.entries(data.source.parameters ?? {})
              .map(([k, v]) => `${k}=${v}`)
              .join('; ') || 'none'
          }
        />
        <Info label="App version" value={data.source.appVersion ?? __APP_VERSION__} />
      </dl>
    </section>
  );
}

function ConfidenceBadge({ confidence }: { confidence?: TopicPeriod['confidence'] }) {
  if (!confidence) {
    return null;
  }
  const title = confidence.evidence.join('\n');
  return (
    <div
      className="mt-2 inline-flex items-center gap-1 rounded-md border border-ink/10 px-2 py-1 text-xs font-bold text-ink/60"
      title={title}
    >
      <CircleHelp className="h-3.5 w-3.5" aria-hidden="true" />
      {confidence.level} {Math.round(confidence.score * 100)}%
    </div>
  );
}

function Timeline({ topics }: { topics: TopicPeriod[] }) {
  if (topics.length === 0) {
    return <EmptyState title="No matching timeline periods" />;
  }
  return (
    <div className="grid gap-3">
      {topics.map((topic) => (
        <article
          key={topic.id}
          className="grid gap-4 rounded-md border border-ink/10 bg-white p-4 md:grid-cols-[150px_minmax(0,1fr)]"
        >
          <div>
            <div className="text-sm font-bold text-rust">{formatMonth(topic.start)}</div>
            <div className="mt-1 text-xs text-ink/50">{topic.messageCount} messages</div>
          </div>
          <div className="min-w-0">
            <h2 className="text-xl font-black text-ink">{topic.label}</h2>
            <ConfidenceBadge confidence={topic.confidence} />
            <p className="mt-2 text-sm leading-6 text-ink/70">{topic.summary}</p>
            <div className="mt-3 flex flex-wrap gap-2">
              {topic.keywords.map((keyword) => (
                <span
                  key={keyword}
                  className="rounded-md bg-wheat/35 px-2 py-1 text-xs font-semibold text-ink"
                >
                  {keyword}
                </span>
              ))}
            </div>
          </div>
        </article>
      ))}
    </div>
  );
}

function MapView({ data, introductions }: { data: Dashboard; introductions: Introduction[] }) {
  return (
    <div className="grid gap-5">
      <section className="rounded-md border border-ink/10 bg-white p-4">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="flex items-center gap-2 text-2xl font-black">
              <Network className="h-6 w-6 text-lake" aria-hidden="true" />
              Who Introduced Whom
            </h2>
            <p className="mt-1 text-sm text-ink/65">
              First pre-arrival mentions become edges in the group memory map.
            </p>
          </div>
          <a
            className="inline-flex h-10 items-center gap-2 rounded-md bg-moss px-4 text-sm font-bold text-white hover:bg-ink"
            href={data.repositoryUrl}
            target="_blank"
            rel="noreferrer"
          >
            <Github className="h-4 w-4" aria-hidden="true" />
            Star this map
          </a>
        </div>
        <div className="mt-4 overflow-auto rounded-md border border-ink/10 bg-[#fbfaf6] p-3">
          <img
            src={assetUrl(data.graph.svgPath)}
            alt="GraphViz map showing who introduced whom in the group chat"
            className="mx-auto min-h-[260px] w-full min-w-[680px] object-contain"
          />
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-2">
        {introductions.map((edge) => (
          <article key={`${edge.from}-${edge.to}`} className="rounded-md border border-ink/10 bg-white p-4">
            <div className="flex items-center gap-2 text-sm font-semibold text-rust">
              <GitFork className="h-4 w-4" aria-hidden="true" />
              {formatDate(edge.firstMentionAt)}
            </div>
            <h3 className="mt-2 text-lg font-black">
              {edge.from} introduced {edge.to}
            </h3>
            <ConfidenceBadge confidence={edge.confidence} />
            <p className="mt-2 text-sm leading-6 text-ink/70">{edge.snippet}</p>
          </article>
        ))}
      </section>
    </div>
  );
}

function Jokes({ jokes }: { jokes: InsideJoke[] }) {
  if (jokes.length === 0) {
    return <EmptyState title="No matching inside-joke origins" />;
  }
  return (
    <div className="grid gap-3 md:grid-cols-2">
      {jokes.map((joke) => (
        <article key={joke.phrase} className="rounded-md border border-ink/10 bg-white p-4">
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-xl font-black capitalize">{joke.phrase}</h2>
            <span className="rounded-md bg-lake/10 px-2 py-1 text-xs font-bold text-lake">
              {joke.occurrences}x
            </span>
          </div>
          <ConfidenceBadge confidence={joke.confidence} />
          <p className="mt-3 text-sm leading-6 text-ink/70">{joke.snippet}</p>
          <div className="mt-4 text-sm text-ink/60">
            Origin: <strong className="text-ink">{joke.originSender}</strong> on {formatDate(joke.originAt)}
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            {joke.participants.map((participant) => (
              <span
                key={participant}
                className="rounded-md bg-rust/10 px-2 py-1 text-xs font-semibold text-rust"
              >
                {participant}
              </span>
            ))}
          </div>
        </article>
      ))}
    </div>
  );
}

function Departures({ departures }: { departures: Departure[] }) {
  if (departures.length === 0) {
    return <EmptyState title="No matching member activity" />;
  }
  return (
    <div className="grid gap-3">
      {departures.map((departure) => (
        <article
          key={departure.member}
          className="grid gap-4 rounded-md border border-ink/10 bg-white p-4 md:grid-cols-[160px_minmax(0,1fr)_130px]"
        >
          <div>
            <h2 className="text-lg font-black">{departure.member}</h2>
            <span
              className={`mt-2 inline-flex rounded-md px-2 py-1 text-xs font-bold ${statusClass(departure.status)}`}
            >
              {departure.status}
            </span>
          </div>
          <div>
            <p className="text-sm leading-6 text-ink/70">{departure.interpretation}</p>
            <ConfidenceBadge confidence={departure.confidence} />
            <p className="mt-2 text-sm italic text-ink/60">{departure.lastSnippet}</p>
          </div>
          <div className="text-left md:text-right">
            <div className="text-2xl font-black tabular-nums">{departure.daysSinceActive}</div>
            <div className="text-xs font-semibold uppercase tracking-wide text-ink/50">days quiet</div>
          </div>
        </article>
      ))}
    </div>
  );
}

function Footer({ data, commit }: { data: Dashboard; commit: string }) {
  return (
    <footer className="border-t border-ink/10 bg-[#fbfaf6] px-4 py-5">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-3 text-sm text-ink/60 md:flex-row md:items-center md:justify-between">
        <div>
          Version <span className="font-mono text-ink">v{__APP_VERSION__}</span> · Commit{' '}
          <span className="font-mono text-ink">{commit}</span>
        </div>
        <div className="flex flex-wrap gap-3">
          <a
            className="font-semibold text-moss hover:text-ink"
            href={data.repositoryUrl}
            target="_blank"
            rel="noreferrer"
          >
            https://github.com/baditaflorin/group-chat-archaeologist
          </a>
          <a
            className="font-semibold text-rust hover:text-ink"
            href={data.paypalUrl}
            target="_blank"
            rel="noreferrer"
          >
            https://www.paypal.com/paypalme/florinbadita
          </a>
        </div>
      </div>
    </footer>
  );
}

function ShellState({ title, body }: { title: string; body: string }) {
  return (
    <main className="grid min-h-screen place-items-center bg-paper px-4 text-ink">
      <section className="max-w-lg rounded-md border border-ink/10 bg-white p-6 shadow-soft">
        <h1 className="text-2xl font-black">{title}</h1>
        <p className="mt-3 text-sm leading-6 text-ink/70">{body}</p>
      </section>
    </main>
  );
}

function EmptyState({ title }: { title: string }) {
  return (
    <div className="rounded-md border border-dashed border-ink/20 bg-white p-8 text-center">
      <h2 className="text-xl font-black">{title}</h2>
    </div>
  );
}

function filterData(data: Dashboard, search: string, members: string[]) {
  const term = search.trim().toLowerCase();
  const memberSet = new Set(members);
  const includesTerm = (value: string) => value.toLowerCase().includes(term);

  return {
    topics: data.topics.filter(
      (topic) =>
        topic.topMembers.some((member) => memberSet.has(member)) &&
        (!term || [topic.label, topic.summary, ...topic.keywords, ...topic.topMembers].some(includesTerm))
    ),
    introductions: data.introductions.filter(
      (edge) =>
        memberSet.has(edge.from) &&
        memberSet.has(edge.to) &&
        (!term || [edge.from, edge.to, edge.snippet].some(includesTerm))
    ),
    insideJokes: data.insideJokes.filter(
      (joke) =>
        joke.participants.some((member) => memberSet.has(member)) &&
        (!term || [joke.phrase, joke.snippet, joke.originSender, ...joke.participants].some(includesTerm))
    ),
    departures: data.departures.filter(
      (departure) =>
        memberSet.has(departure.member) &&
        (!term || [departure.member, departure.status, departure.lastSnippet].some(includesTerm))
    )
  };
}

function statusClass(status: string) {
  if (status === 'departed') {
    return 'bg-rust/10 text-rust';
  }
  if (status === 'quiet') {
    return 'bg-wheat/35 text-ink';
  }
  return 'bg-moss/10 text-moss';
}

function percent(value?: number) {
  if (typeof value !== 'number') {
    return 'n/a';
  }
  return `${Math.round(value * 100)}%`;
}
