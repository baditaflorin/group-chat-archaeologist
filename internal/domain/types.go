package domain

import "time"

const SchemaVersion = "v1"

type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
	Text      string    `json:"text"`
	Source    string    `json:"source,omitempty"`
	Line      int       `json:"line,omitempty"`
}

type Dashboard struct {
	SchemaVersion   string             `json:"schemaVersion"`
	GeneratedAt     string             `json:"generatedAt"`
	RepositoryURL   string             `json:"repositoryUrl"`
	PayPalURL       string             `json:"paypalUrl"`
	Source          SourceSummary      `json:"source"`
	Members         []Member           `json:"members"`
	Topics          []TopicPeriod      `json:"topics"`
	Introductions   []IntroductionEdge `json:"introductions"`
	InsideJokes     []InsideJoke       `json:"insideJokes"`
	Departures      []Departure        `json:"departures"`
	NotableMessages []NotableMessage   `json:"notableMessages"`
	Warnings        []Warning          `json:"warnings"`
	Debug           DebugInfo          `json:"debug"`
	Graph           GraphArtifacts     `json:"graph"`
}

type SourceSummary struct {
	InputName          string            `json:"inputName"`
	InputSHA256        string            `json:"inputSha256"`
	Parser             string            `json:"parser"`
	Adapter            string            `json:"adapter"`
	AdapterConfidence  float64           `json:"adapterConfidence"`
	ExtractionMode     string            `json:"extractionMode"`
	NormalizationSteps []string          `json:"normalizationSteps"`
	AnalyticsEngine    string            `json:"analyticsEngine"`
	MessageCount       int               `json:"messageCount"`
	MemberCount        int               `json:"memberCount"`
	FirstMessageAt     string            `json:"firstMessageAt"`
	LastMessageAt      string            `json:"lastMessageAt"`
	WarningCount       int               `json:"warningCount"`
	LLMProvider        string            `json:"llmProvider"`
	LLMModel           string            `json:"llmModel"`
	LLMUsed            bool              `json:"llmUsed"`
	SourceCommit       string            `json:"sourceCommit"`
	AppVersion         string            `json:"appVersion"`
	Parameters         map[string]string `json:"parameters"`
}

type Member struct {
	Name           string `json:"name"`
	MessageCount   int    `json:"messageCount"`
	FirstMessageAt string `json:"firstMessageAt"`
	LastMessageAt  string `json:"lastMessageAt"`
}

type TopicPeriod struct {
	ID           string     `json:"id"`
	Label        string     `json:"label"`
	Start        string     `json:"start"`
	End          string     `json:"end"`
	MessageCount int        `json:"messageCount"`
	Keywords     []string   `json:"keywords"`
	TopMembers   []string   `json:"topMembers"`
	Summary      string     `json:"summary"`
	Confidence   Confidence `json:"confidence"`
}

type IntroductionEdge struct {
	From           string     `json:"from"`
	To             string     `json:"to"`
	FirstMentionAt string     `json:"firstMentionAt"`
	MessageID      string     `json:"messageId"`
	Snippet        string     `json:"snippet"`
	Confidence     Confidence `json:"confidence"`
}

type InsideJoke struct {
	Phrase       string     `json:"phrase"`
	OriginAt     string     `json:"originAt"`
	OriginSender string     `json:"originSender"`
	OriginID     string     `json:"originId"`
	Occurrences  int        `json:"occurrences"`
	Participants []string   `json:"participants"`
	Snippet      string     `json:"snippet"`
	Confidence   Confidence `json:"confidence"`
}

type Departure struct {
	Member          string     `json:"member"`
	Status          string     `json:"status"`
	LastMessageAt   string     `json:"lastMessageAt"`
	DaysSinceActive int        `json:"daysSinceActive"`
	ActiveSpanDays  int        `json:"activeSpanDays"`
	LastSnippet     string     `json:"lastSnippet"`
	Interpretation  string     `json:"interpretation"`
	Confidence      Confidence `json:"confidence"`
}

type NotableMessage struct {
	ID      string `json:"id"`
	At      string `json:"at"`
	Sender  string `json:"sender"`
	Kind    string `json:"kind"`
	Snippet string `json:"snippet"`
	Why     string `json:"why"`
}

type GraphArtifacts struct {
	DOTPath     string `json:"dotPath"`
	SVGPath     string `json:"svgPath"`
	Rendered    bool   `json:"rendered"`
	Renderer    string `json:"renderer"`
	RenderError string `json:"renderError,omitempty"`
}

type Confidence struct {
	Score    float64  `json:"score"`
	Level    string   `json:"level"`
	Evidence []string `json:"evidence"`
}

type Warning struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Why      string `json:"why"`
	NextStep string `json:"nextStep"`
	Line     int    `json:"line,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

type DebugInfo struct {
	AdapterEvidence []string `json:"adapterEvidence"`
	ParseWarnings   int      `json:"parseWarnings"`
}

type MemberStat struct {
	Name           string
	MessageCount   int
	FirstMessageAt time.Time
	LastMessageAt  time.Time
}

type StorageSummary struct {
	MemberStats []MemberStat
	Engine      string
}
