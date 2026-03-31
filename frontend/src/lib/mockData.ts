import {
  APIKey,
  AppConfig,
  CatalogCourse,
  CourseSection,
  CronLog,
  DailyStat,
  InstanceState,
  PromotionLog,
  RequestEvent,
  ScraperLog,
  User
} from "../types/models";

export const mockUsers: User[] = [
  {
    uid: "u_001",
    email: "allen.zheng@utdallas.edu",
    displayName: "Allen Zheng",
    photoURL: "https://storage.googleapis.com/acm-officer-database.firebasestorage.app/officers/F6vG5bnAUjctcMtxKQ6nYMPYntw2",
    isAdmin: false,
    approvalStatus: "approved",
    createdAt: "2026-01-10T15:24:00.000Z",
    lastLoginAt: "2026-03-30T19:30:00.000Z"
  },
  {
    uid: "u_002",
    email: "rei-shibatani@acmutd.co",
    displayName: "Rei Shibatani",
    photoURL: "https://storage.googleapis.com/acm-officer-database.firebasestorage.app/officers/W3NS6yRINrRiFmSD84NKu4rHRGw1",
    isAdmin: true,
    approvalStatus: "approved",
    createdAt: "2025-09-08T11:10:00.000Z",
    lastLoginAt: "2026-03-31T13:06:00.000Z"
  }
];

export const mockApiKeys: APIKey[] = [
  {
    keyId: "key_1",
    key: "acm_live_ce3345_a1b2c3d4e5",
    userId: "u_001",
    ownerEmail: "allen.zheng@utdallas.edu",
    label: "HackUTD Team API",
    description: "Project Spark planner backend",
    status: "active",
    isAdmin: false,
    rateLimit: 120,
    windowSeconds: 60,
    createdAt: "2026-03-01T08:20:00.000Z",
    expiresAt: "2026-12-31T23:59:59.000Z",
    usageCount: 1234,
    lastUsedAt: "2026-03-31T15:30:00.000Z"
  },
  {
    keyId: "key_2",
    key: "acm_admin_ops_z1y2x3w4v5",
    userId: "u_002",
    ownerEmail: "rei-shibatani@acmutd.co",
    label: "Admin Ops Key",
    description: "Internal ACM automation jobs",
    status: "active",
    isAdmin: true,
    rateLimit: 1000,
    windowSeconds: 60,
    createdAt: "2026-01-15T12:10:00.000Z",
    expiresAt: "2027-01-15T12:10:00.000Z",
    usageCount: 9901,
    lastUsedAt: "2026-03-31T16:02:00.000Z"
  }
];

export const mockDailyStats: DailyStat[] = [
  {
    statId: "key_1_2026-03-25",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-25",
    totalRequests: 182,
    successCount: 176,
    errorCount: 6,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 83 },
      { endpoint: "/grades/CE/3345", count: 62 },
      { endpoint: "/professors", count: 37 }
    ]
  },
  {
    statId: "key_1_2026-03-26",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-26",
    totalRequests: 210,
    successCount: 204,
    errorCount: 6,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 94 },
      { endpoint: "/grades/CE/3345", count: 66 },
      { endpoint: "/professors", count: 50 }
    ]
  },
  {
    statId: "key_1_2026-03-27",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-27",
    totalRequests: 165,
    successCount: 161,
    errorCount: 4,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 71 },
      { endpoint: "/grades/CE/3345", count: 55 },
      { endpoint: "/professors", count: 39 }
    ]
  },
  {
    statId: "key_1_2026-03-28",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-28",
    totalRequests: 246,
    successCount: 237,
    errorCount: 9,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 112 },
      { endpoint: "/grades/CE/3345", count: 81 },
      { endpoint: "/professors", count: 53 }
    ]
  },
  {
    statId: "key_1_2026-03-29",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-29",
    totalRequests: 195,
    successCount: 190,
    errorCount: 5,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 88 },
      { endpoint: "/grades/CE/3345", count: 64 },
      { endpoint: "/professors", count: 43 }
    ]
  },
  {
    statId: "key_1_2026-03-30",
    keyId: "key_1",
    userId: "u_001",
    date: "2026-03-30",
    totalRequests: 236,
    successCount: 229,
    errorCount: 7,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 108 },
      { endpoint: "/grades/CE/3345", count: 71 },
      { endpoint: "/professors", count: 57 }
    ]
  },
  {
    statId: "key_1_2026-03-31",
    keyId: "key_2",
    userId: "u_002",
    date: "2026-03-31",
    totalRequests: 222,
    successCount: 214,
    errorCount: 8,
    endpointBreakdown: [
      { endpoint: "/courses/CE/3345/sections", count: 102 },
      { endpoint: "/grades/CE/3345", count: 69 },
      { endpoint: "/professors", count: 51 }
    ]
  }
];

export const mockRecentRequests: RequestEvent[] = [
  { id: "r_1", userId: "u_001", keyId: "key_1", dateTime: "2026-03-31T16:01:22.000Z", endpoint: "/courses/CE/3345/sections", method: "GET", statusCode: 200, latencyMs: 118 },
  { id: "r_2", userId: "u_001", keyId: "key_1", dateTime: "2026-03-31T15:55:04.000Z", endpoint: "/grades/CE/3345", method: "GET", statusCode: 200, latencyMs: 154 },
  { id: "r_3", userId: "u_001", keyId: "key_1", dateTime: "2026-03-31T15:45:18.000Z", endpoint: "/courses/CS/1337/sections", method: "GET", statusCode: 200, latencyMs: 102 },
  { id: "r_4", userId: "u_001", keyId: "key_1", dateTime: "2026-03-31T15:21:33.000Z", endpoint: "/professors", method: "GET", statusCode: 429, latencyMs: 61 },
  { id: "r_5", userId: "u_001", keyId: "key_1", dateTime: "2026-03-31T15:12:12.000Z", endpoint: "/courses/CE/3345/sections", method: "GET", statusCode: 200, latencyMs: 128 },
  { id: "r_6", userId: "u_002", keyId: "key_2", dateTime: "2026-03-31T14:59:00.000Z", endpoint: "/courses/CS/3341/sections", method: "GET", statusCode: 200, latencyMs: 141 },
  { id: "r_7", userId: "u_002", keyId: "key_2", dateTime: "2026-03-31T14:47:42.000Z", endpoint: "/grades/CS/3341", method: "GET", statusCode: 500, latencyMs: 210 },
  { id: "r_8", userId: "u_002", keyId: "key_2", dateTime: "2026-03-31T14:33:11.000Z", endpoint: "/courses/CE/3345/sections", method: "GET", statusCode: 200, latencyMs: 107 },
  { id: "r_9", userId: "u_002", keyId: "key_2", dateTime: "2026-03-31T14:20:49.000Z", endpoint: "/courses/CE/3345/sections", method: "GET", statusCode: 200, latencyMs: 113 },
  { id: "r_10", userId: "u_002", keyId: "key_2", dateTime: "2026-03-31T14:11:27.000Z", endpoint: "/health", method: "GET", statusCode: 200, latencyMs: 22 }
];

export const mockAppConfig: AppConfig = {
  autoApproveMode: "utdallas",
  hackutdModeEnabled: true,
  hackutdEndDate: "2026-11-03T23:59:59.000Z",
  currentSemester: "26s",
  instanceType: "t3.large",
  updatedAt: "2026-03-31T13:00:00.000Z",
  updatedBy: "u_002"
};

export const mockScraperLogs: ScraperLog[] = [
  {
    logId: "scr_001",
    startedAt: "2026-03-31T01:00:00.000Z",
    completedAt: "2026-03-31T01:06:24.000Z",
    status: "success",
    semester: "26s",
    coursesWritten: 5487,
    sectionsWritten: 19641,
    mergeMatchRate: 0.94,
    triggeredBy: "cron"
  },
  {
    logId: "scr_002",
    startedAt: "2026-03-30T01:00:00.000Z",
    completedAt: "2026-03-30T01:08:30.000Z",
    status: "error",
    semester: "26s",
    coursesWritten: 5250,
    sectionsWritten: 17840,
    mergeMatchRate: 0.89,
    triggeredBy: "admin:u_002",
    errorMessage: "RMP profile merge timed out"
  },
  {
    logId: "scr_003",
    startedAt: "2026-03-29T01:00:00.000Z",
    completedAt: "2026-03-29T01:06:04.000Z",
    status: "success",
    semester: "26s",
    coursesWritten: 5478,
    sectionsWritten: 19610,
    mergeMatchRate: 0.93,
    triggeredBy: "cron"
  }
];

export const mockCronLogs: CronLog[] = [
  {
    date: "2026-03-31",
    runAt: "2026-03-31T09:00:00.000Z",
    tokensChecked: 412,
    validTokenCount: 376,
    action: "started",
    instanceState: "running",
    notes: "HackUTD weekday auto-start"
  },
  {
    date: "2026-03-30",
    runAt: "2026-03-30T23:30:00.000Z",
    tokensChecked: 391,
    validTokenCount: 360,
    action: "stopped",
    instanceState: "stopped",
    notes: "Outside active usage window"
  },
  {
    date: "2026-03-29",
    runAt: "2026-03-29T09:00:00.000Z",
    tokensChecked: 405,
    validTokenCount: 367,
    action: "no-op",
    instanceState: "running",
    notes: "Server already running"
  }
];

export const mockPromotionLogs: PromotionLog[] = [
  {
    logId: "promo_001",
    promotedAt: "2026-03-27T17:05:00.000Z",
    documentCount: 80432,
    promotedBy: "rei-shibatani@acmutd.co",
    fromEnv: "dev",
    toEnv: "prod",
    notes: "Post-validation promotion"
  },
  {
    logId: "promo_002",
    promotedAt: "2026-03-20T18:20:00.000Z",
    documentCount: 79910,
    promotedBy: "rei-shibatani@acmutd.co",
    fromEnv: "dev",
    toEnv: "prod",
    notes: "Weekly sync"
  }
];

export const mockCatalogCourses: CatalogCourse[] = [
  {
    prefix: "CE",
    number: "3345",
    title: "Data Structures and Introduction to Algorithmic Analysis",
    description: "Study of abstract data types, sorting, searching and asymptotic analysis.",
    classLevel: "Undergraduate",
    school: "Engineering and Computer Science",
    schoolId: "ECS",
    aggregatedGrades: { A: 112, B: 84, C: 33, D: 9, F: 4 },
    sectionTypes: ["Lecture", "Lab"],
    crossListed: []
  }
];

export const mockCourseSections: CourseSection[] = [
  {
    prefix: "CE",
    number: "3345",
    section: "001",
    term: "26s",
    sectionAddress: "ce3345.001.26s",
    title: "Data Structures",
    description: "Implementation and analysis of classic data structures.",
    enrolledStatus: "Open",
    enrolledCurrent: 93,
    enrolledMax: 100,
    waitlist: 0,
    startDate: "2026-01-20",
    endDate: "2026-05-08",
    meetings: [
      {
        dateRange: "January 20, 2026-May 8, 2026",
        days: ["Monday", "Wednesday"],
        time: "10:00am-11:15am",
        location: "ECSS 2.415"
      }
    ],
    activityType: "Lecture",
    semesterCreditHours: "3",
    grading: "Letter",
    sessionType: "Regular Academic Session",
    school: "Engineering and Computer Science",
    schoolId: "ECS",
    enrollmentReqs: ["CS 1337 minimum grade C"],
    classAttributes: ["Core Communication"],
    classNotes: null,
    instructors: ["Prof. Jane Park"],
    instructorIds: ["inst_001"],
    crossListed: ["CS 3345"],
    syllabus: "https://example.com/syllabus/ce3345",
    gradeDistribution: { A: 46, "A-": 20, "B+": 14, B: 9, "C+": 4, C: 3, D: 1, F: 1 }
  }
];

export const mockInstanceState: InstanceState = {
  instanceId: "i-0a1b2c3d4e5f67890",
  instanceType: "t3.large",
  state: "running",
  uptimeSeconds: 12060,
  publicIp: "34.72.188.19"
};
