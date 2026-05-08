import { z } from 'zod';

export const memberSchema = z.object({
  name: z.string(),
  messageCount: z.number(),
  firstMessageAt: z.string(),
  lastMessageAt: z.string()
});

export const topicSchema = z.object({
  id: z.string(),
  label: z.string(),
  start: z.string(),
  end: z.string(),
  messageCount: z.number(),
  keywords: z.array(z.string()),
  topMembers: z.array(z.string()),
  summary: z.string()
});

export const introductionSchema = z.object({
  from: z.string(),
  to: z.string(),
  firstMentionAt: z.string(),
  messageId: z.string(),
  snippet: z.string()
});

export const insideJokeSchema = z.object({
  phrase: z.string(),
  originAt: z.string(),
  originSender: z.string(),
  originId: z.string(),
  occurrences: z.number(),
  participants: z.array(z.string()),
  snippet: z.string()
});

export const departureSchema = z.object({
  member: z.string(),
  status: z.string(),
  lastMessageAt: z.string(),
  daysSinceActive: z.number(),
  activeSpanDays: z.number(),
  lastSnippet: z.string(),
  interpretation: z.string()
});

export const notableMessageSchema = z.object({
  id: z.string(),
  at: z.string(),
  sender: z.string(),
  kind: z.string(),
  snippet: z.string(),
  why: z.string()
});

export const dashboardSchema = z.object({
  schemaVersion: z.string(),
  generatedAt: z.string(),
  repositoryUrl: z.string().url(),
  paypalUrl: z.string().url(),
  source: z.object({
    inputName: z.string(),
    inputSha256: z.string(),
    parser: z.string(),
    extractionMode: z.string(),
    analyticsEngine: z.string(),
    messageCount: z.number(),
    memberCount: z.number(),
    firstMessageAt: z.string(),
    lastMessageAt: z.string(),
    llmProvider: z.string(),
    llmModel: z.string(),
    llmUsed: z.boolean(),
    sourceCommit: z.string()
  }),
  members: z.array(memberSchema),
  topics: z.array(topicSchema),
  introductions: z.array(introductionSchema),
  insideJokes: z.array(insideJokeSchema),
  departures: z.array(departureSchema),
  notableMessages: z.array(notableMessageSchema),
  graph: z.object({
    dotPath: z.string(),
    svgPath: z.string(),
    rendered: z.boolean(),
    renderer: z.string(),
    renderError: z.string().optional()
  })
});

export const metaSchema = z.object({
  generatedAt: z.string(),
  schemaVersion: z.string(),
  sourceCommit: z.string(),
  inputSha256: z.string(),
  messageCount: z.number(),
  graphRendered: z.boolean(),
  artifactVersion: z.string()
});

export type Dashboard = z.infer<typeof dashboardSchema>;
export type TopicPeriod = z.infer<typeof topicSchema>;
export type Introduction = z.infer<typeof introductionSchema>;
export type InsideJoke = z.infer<typeof insideJokeSchema>;
export type Departure = z.infer<typeof departureSchema>;
export type Meta = z.infer<typeof metaSchema>;
