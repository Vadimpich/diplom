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
  user: User;
}

export interface Specialist {
  id: number;
  full_name: string;
  personnel_number: string | null;
  created_at: string;
  updated_at: string;
}

export interface SpecialistsResponse {
  items: Specialist[];
}

export type ExaminationStatus =
  | "created"
  | "collecting_answers"
  | "ready_for_processing";

export interface Examination {
  id: number;
  specialist_id: number;
  created_by_user_id: number;
  questionnaire_id: number | null;
  status: ExaminationStatus;
  created_at: string;
  started_at: string | null;
  finished_at: string | null;
  updated_at: string;
}

export interface ExaminationsResponse {
  items: Examination[];
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
  questions: Question[];
  created_at: string;
  updated_at: string;
}

export interface QuestionnairesResponse {
  items: Questionnaire[];
}

export interface HealthResponse {
  status: "ok" | "degraded";
  database: "up" | "down";
}

export interface ApiErrorShape {
  message?: string;
  error?: string;
  details?: string;
}
