import type { Dashboard } from './schema';

export function filterDashboard(data: Dashboard, search: string, members: string[]) {
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
