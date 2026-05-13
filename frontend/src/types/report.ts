export interface HeadlineMetric {
  title: string
  value: string
  delta: string
  insight: string
}

export interface ComparisonRow {
  dimension: string
  old: string
  new: string
  impact: string
}

export interface TrendPoint {
  name: string
  old: number
  new: number
  delta: number
}

export interface SectionHighlights {
  overview: string
  key_changes: string
  impacts: string
  actions: string
}

export interface SectionBodies {
  overview: string
  key_changes: string
  impacts: string
  actions: string
}

// 五法解读类型定义
export interface WordingChange {
  topic: string
  old: string
  new: string
  signal: string
}

export interface RankingChanges {
  rising: string[]
  falling: string[]
  new: string[]
  disappeared: string[]
}

export interface NewPhrase {
  phrase: string
  meaning: string
  impact: string
}

export interface DisappearedPhrase {
  phrase: string
  context: string
  signal: string
}

export interface FocusShift {
  macro_tone: string
  priority_directions: string[]
  risk_areas: string[]
}

export interface StructuredReport {
  executive_summary: string
  headline_metrics: HeadlineMetric[]
  comparison_table: ComparisonRow[]
  trend_points: TrendPoint[]
  section_highlights: SectionHighlights
  sections: SectionBodies
  // 五法解读字段
  wording_changes?: WordingChange[]
  ranking_changes?: RankingChanges
  new_phrases?: NewPhrase[]
  disappeared_phrases?: DisappearedPhrase[]
  focus_shift?: FocusShift
}
