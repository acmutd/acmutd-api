export interface User {
  uid: string;
  email: string;
  displayName: string;
  photoURL?: string;
  isAdmin: boolean;
  approvalStatus: "approved" | "pending" | "banned";
  createdAt: string;
  lastLoginAt: string;
}

export type APIKeyStatus = "active" | "inactive" | "pending" | "rejected";

export interface APIKey {
  keyId: string;
  key: string;
  userId: string;
  ownerEmail: string;
  label: string;
  description: string;
  status: APIKeyStatus;
  isAdmin: boolean;
  rateLimit: number;
  windowSeconds: number;
  createdAt: string;
  expiresAt: string;
  usageCount: number;
  lastUsedAt: string | null;
}

export interface EndpointBreakdown {
  endpoint: string;
  count: number;
}

export interface DailyStat {
  statId: string;
  keyId: string;
  userId: string;
  date: string;
  totalRequests: number;
  successCount: number;
  errorCount: number;
  endpointBreakdown: EndpointBreakdown[];
}

export type AutoApproveMode = "none" | "utdallas" | "acmutd" | "all";

export interface AppConfig {
  autoApproveMode: AutoApproveMode;
  hackutdModeEnabled: boolean;
  hackutdEndDate: string;
  currentSemester: string;
  instanceType: "t3.micro" | "t3.small" | "t3.large";
  updatedAt: string;
  updatedBy: string;
}

export interface ScraperLog {
  logId: string;
  startedAt: string;
  completedAt: string;
  status: "running" | "success" | "error";
  semester: string;
  coursesWritten: number;
  sectionsWritten: number;
  mergeMatchRate: number;
  triggeredBy: string;
  errorMessage?: string;
}

export interface CronLog {
  date: string;
  runAt: string;
  tokensChecked: number;
  validTokenCount: number;
  action: "started" | "stopped" | "no-op";
  instanceState: string;
  notes?: string;
}

export interface PromotionLog {
  logId: string;
  promotedAt: string;
  documentCount: number;
  promotedBy: string;
  fromEnv: "dev" | "prod";
  toEnv: "dev" | "prod";
  notes?: string;
}

export interface CatalogCourse {
  prefix: string;
  number: string;
  title: string;
  description: string;
  classLevel: string;
  school: string;
  schoolId: string;
  aggregatedGrades?: Record<string, number>;
  sectionTypes: string[];
  crossListed: string[];
}

export interface SectionMeeting {
  dateRange: string;
  days: string[];
  time: string;
  location: string;
}

export interface CourseSection {
  prefix: string;
  number: string;
  section: string;
  term: string;
  sectionAddress: string;
  title: string;
  description: string;
  enrolledStatus: string;
  enrolledCurrent: number;
  enrolledMax: number;
  waitlist: number;
  startDate: string;
  endDate: string;
  meetings: SectionMeeting[];
  activityType: string;
  semesterCreditHours: string;
  grading: string;
  sessionType: string;
  school: string;
  schoolId: string;
  enrollmentReqs: string[];
  classAttributes: string[];
  classNotes?: string | null;
  instructors: string[];
  instructorIds: string[];
  crossListed: string[];
  syllabus?: string;
  gradeDistribution?: Record<string, number> | null;
}

export interface RequestEvent {
  id: string;
  userId: string;
  keyId: string;
  dateTime: string;
  endpoint: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  statusCode: number;
  latencyMs: number;
}

export interface InstanceState {
  instanceId: string;
  instanceType: "t3.micro" | "t3.small" | "t3.large";
  state: "running" | "stopped";
  uptimeSeconds: number;
  publicIp: string;
}
