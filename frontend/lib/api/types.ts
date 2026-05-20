export type RoleSlug = "admin" | "operator";

export interface Role {
  id: number;
  slug: RoleSlug;
  name: string;
}

export interface User {
  id: number;
  login: string;
  role: Role;
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UsersResponse {
  items: User[];
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
  refresh_expires_in: number;
  user: User;
}

export interface Specialist {
  id: number;
  full_name: string;
  personnel_number: string | null;
  examinations_count: number;
  last_examination_id: number | null;
  last_examination_at: string | null;
  last_examination_status: ExaminationStatus | null;
  last_overall_score: number | null;
  last_overall_band: string | null;
  baseline_exam_count: number;
  baseline_refreshed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface SpecialistsResponse {
  items: Specialist[];
}

export type ExaminationStatus =
  | "created"
  | "collecting_answers"
  | "ready_for_processing"
  | "processing"
  | "aggregating"
  | "aggregated"
  | "decision_pending"
  | "completed"
  | "failed";

export type ProcessingChannelName = "text" | "acoustic" | "paralinguistic";

export type ProcessingChannelStatus =
  | "queued"
  | "processing"
  | "succeeded"
  | "failed_temporary"
  | "failed_fatal"
  | "temporary_error"
  | "fatal_error"
  | "exhausted"
  | "retry_scheduled";

export interface Examination {
  id: number;
  specialist_id: number;
  created_by_user_id: number;
  questionnaire_id: number | null;
  questions?: ExaminationQuestionSnapshot[];
  status: ExaminationStatus;
  created_at: string;
  started_at: string | null;
  finished_at: string | null;
  updated_at: string;
}

export interface ExaminationQuestionSnapshot {
  id: number;
  examination_id: number;
  specialist_id: number;
  questionnaire_id: number;
  source_question_id?: number | null;
  position: number;
  text: string;
}

export interface ExaminationsResponse {
  items: Examination[];
}

export interface ExaminationProcessingStatusChannel {
  channel: ProcessingChannelName;
  status: ProcessingChannelStatus;
  attempt_count: number;
  max_attempts: number;
  message_version: number;
  queued_at: string | null;
  started_at: string | null;
  finished_at: string | null;
  last_error_code: string | null;
  last_error_message: string | null;
  broker_message_id: string | null;
  broker_correlation_id: string | null;
}

export interface ExaminationProcessingStatus {
  examination_id: number;
  status: Extract<
    ExaminationStatus,
    "ready_for_processing" | "processing" | "aggregating" | "aggregated" | "decision_pending" | "completed" | "failed"
  >;
  message_version: number;
  channels_total: number;
  channels_completed: number;
  terminal: boolean;
  started_at: string | null;
  updated_at: string;
  finished_at: string | null;
  failed_at: string | null;
  channels: ExaminationProcessingStatusChannel[];
}

export interface Answer {
  id: number;
  examination_id: number;
  examination_question_id: number;
  specialist_id: number;
  created_by_user_id: number;
  text: string;
  audio_s3_key: string;
  created_at: string;
}

export interface Question {
  id: number;
  text: string;
  position: number;
}

export interface Questionnaire {
  id: number;
  title: string;
  description: string | null;
  is_active: boolean;
  usage_count: number;
  last_used_at: string | null;
  last_edited_at: string;
  last_editor: QuestionnaireEditor | null;
  questions: Question[];
  created_at: string;
  updated_at: string;
}

export interface QuestionnaireEditor {
  id: number;
  login: string;
}

export interface QuestionnairesResponse {
  items: Questionnaire[];
}

export interface HealthResponse {
  status: "ok" | "degraded";
  service: string;
}

export interface FrontendDependencyState {
  core_backend: "up" | "down";
}

export interface FrontendReadinessResponse {
  status: "ready" | "degraded";
  service: "frontend";
  dependencies: FrontendDependencyState;
}

export interface AdminMonitoringMetrics {
  frontend_dependency_up: 0 | 1;
}

export type AuditOutcome = "succeeded" | "failed" | "rejected";

export interface AuditActor {
  user_id?: number | null;
  login: string;
  role_slug: string;
  ip: string;
  user_agent: string;
}

export interface AuditResourceRef {
  kind: string;
  id: number;
}

export interface AuditDomainRefs {
  examination_id?: number | null;
  specialist_id?: number | null;
  questionnaire_id?: number | null;
  channel?: string | null;
  decision_snapshot_id?: number | null;
}

export interface AuditEvent {
  id: number;
  event_type: string;
  event_key: string;
  outcome: AuditOutcome;
  happened_at: string;
  request_id: string;
  trace_id: string;
  traceparent: string;
  tracestate: string;
  correlation_id: string;
  actor: AuditActor | null;
  resource: AuditResourceRef | null;
  domain_refs: AuditDomainRefs;
  payload: Record<string, unknown>;
}

export interface AuditEventsResponse {
  items: AuditEvent[];
}

export interface AuditEventsQuery {
  event_type?: string;
  resource_kind?: string;
  resource_id?: number;
  from?: string;
  to?: string;
  limit?: number;
}

export interface SystemSettings {
  audio_retention_ttl_days: number;
  processing_max_attempts: number;
  kesmi_max_retries: number;
  created_at: string;
  updated_at: string;
}

export interface ApiErrorShape {
  message?: string;
  error?: string;
  details?: string;
}

export interface ResultSummary {
  overall_score: number;
  overall_band: string;
  primary_metric_key: string;
  neutral_recommendation_placeholder: string;
}

export interface ResultMetric {
  key: string;
  label: string;
  value: number;
  scale: string;
  direction: string;
}

export interface ResultChannelContribution {
  channel: ProcessingChannelName;
  metric_key: string;
  weight: number;
  contribution: number;
  evidence_keys: string[];
}

export interface ResultChannelReportScore {
  key: string;
  label: string;
  value: number;
}

export interface ResultChannelReport {
  channel: ProcessingChannelName;
  model_version: string;
  quality_flags: string[] | null;
  evidence: string[] | null;
  scores: ResultChannelReportScore[];
}

export interface ResultExplanation {
  position: number;
  kind: string;
  text: string;
}

export interface ResultBaselineDeviation {
  delta: number;
  band: string;
  baseline_available?: boolean;
  baseline_source?: string;
  reference_population_version?: string;
  baseline_exam_count?: number;
  update_eligible?: boolean;
  data_reliability?: number;
}

export interface ExaminationResult {
  schema_version: number;
  aggregation_version: string;
  examination_id: number;
  specialist_id: number;
  status: Extract<ExaminationStatus, "aggregated" | "aggregating" | "decision_pending" | "completed">;
  generated_at: string;
  summary: ResultSummary;
  metrics: ResultMetric[];
  channel_contributions: ResultChannelContribution[];
  channel_reports: ResultChannelReport[];
  explanations: ResultExplanation[];
  baseline_snapshot: {
    algorithm_version: string;
    refreshed_at: string;
    general: ResultBaselineDeviation;
    personal: ResultBaselineDeviation;
  };
  decision: {
    state: "pending" | "succeeded" | "transport_exhausted" | "business_error";
    recommendation: "unavailable" | "allowed" | "risk" | "denied";
    message: string;
    decision_code?: "allow" | "monitoring" | "extended_check" | "no_access";
    risk_class?: string;
    patterns?: string[];
    correlation_id: string;
    attempt_count: number;
    max_attempts: number;
    last_attempt_at: string | null;
    diagnostics: {
      error_class?: string | null;
      error_code?: string | null;
      error_message?: string | null;
      http_status?: number | null;
      retryable: boolean;
    };
    raw_response_available: boolean;
  };
}

export interface SpecialistResultHistoryMetric {
  key: string;
  label: string;
  value: number;
  previous_value?: number | null;
  delta_from_previous?: number | null;
}

export interface SpecialistResultHistoryItem {
  examination_id: number;
  generated_at: string;
  status: Extract<ExaminationStatus, "aggregated">;
  summary: {
    overall_score: number;
    overall_band: string;
  };
  baseline_snapshot: {
    algorithm_version: string;
    refreshed_at: string;
    general_delta: number;
    personal_delta: number;
    baseline_exam_count: number;
  };
  key_metrics: SpecialistResultHistoryMetric[];
}

export interface SpecialistResultHistoryResponse {
  specialist_id: number;
  items: SpecialistResultHistoryItem[];
}
