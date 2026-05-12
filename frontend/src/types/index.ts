export interface County {
  county_id: string;
  county_name: string;
  state: string;
  url: string;
  status: 'active' | 'broken' | 'paused';
  source_type: string;
  permit_count: number;
  last_run_at: string | null;
  last_run_status?: 'success' | 'failed' | 'running' | null;
  last_run_inserted: number;
  last_generated_at: string | null;
}

export interface DashboardStats {
  total_counties: number;
  active_counties: number;
  broken_counties: number;
  paused_counties: number;
  permits_this_week: number;
  total_permits: number;
  last_run_at: string | null;
  runs_this_week: number;
  error_rate: number;
  by_status: { status: string; count: number }[];
  permits_by_day: { day: string; count: number }[];
  recent_errors: {
    county_id: string;
    error_message: string;
    started_at: string;
  }[];
}

export interface ScrapeRun {
  id: number;
  county_id: string;
  status: 'running' | 'success' | 'failed';
  records_found: number;
  records_inserted: number;
  started_at: string;
  finished_at: string | null;
  error_message: string | null;
}

export interface Permit {
  id: number;
  permit_number: string;
  scraped_at: string;
  raw_data: Record<string, unknown>;
  [key: string]: unknown;
}

export interface PermitsResponse {
  permits: Permit[];
  total: number;
  page: number;
  per_page: number;
}

export interface ErrorRateRow {
  day: string;
  error_pct: number;
  total: number;
  failed: number;
}
