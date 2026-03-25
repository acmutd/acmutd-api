export const debugAdminDashboardData = {
  user: {
    uid: "debug-admin-001",
    name: "Priya Officer",
    email: "officer@utdallas.edu",
    is_admin: true
  },
  auth: {
    oauth_provider: "google.com",
    admin_gate_source: "firestore:isAdmin",
    admin_gate_status: "granted",
    domain_restriction: "utdallas.edu"
  },
  approval: {
    mode: "manual",
    global_toggle_hint: "New key requests require officer approval when manual mode is active"
  },
  requests: [
    {
      uid: "debug-user-001",
      email: "student1@utdallas.edu",
      requested_at: "2026-03-25T18:30:00.000Z",
      request_type: "new",
      status: "pending",
      label: "Senior design API",
      description: "Need course/professor data for capstone dashboard",
      requested_key_type: "standard"
    },
    {
      uid: "debug-user-002",
      email: "student2@utdallas.edu",
      requested_at: "2026-03-25T17:15:00.000Z",
      request_type: "regenerate",
      status: "pending",
      label: "HackUTD team service",
      description: "Token leaked in logs; requesting urgent rotation",
      requested_key_type: "standard"
    }
  ],
  keys: [
    {
      uid: "debug-user-010-key-a",
      email: "activeuser@utdallas.edu",
      label: "Course recommendation microservice",
      type: "standard",
      masked_key: "acm_****_debug",
      status: "active",
      usage_count: 42,
      created_at: "2026-01-10T13:00:00.000Z",
      expires_at: "2026-05-14T23:59:59.000Z",
      last_used_at: "2026-03-25T19:05:00.000Z",
      updated_at: "2026-03-25T19:05:00.000Z"
    },
    {
      uid: "debug-user-011-key-a",
      email: "inactiveuser@utdallas.edu",
      label: "Old analytics script",
      type: "standard",
      masked_key: "acm_****_old",
      status: "inactive",
      usage_count: 3,
      created_at: "2025-09-05T08:00:00.000Z",
      expires_at: "2026-01-15T23:59:59.000Z",
      last_used_at: "2026-01-03T08:10:00.000Z",
      updated_at: "2026-03-24T08:10:00.000Z"
    },
    {
      uid: "debug-admin-001-key-admin",
      email: "officer@utdallas.edu",
      label: "Officer emergency admin key",
      type: "admin",
      masked_key: "acm_****_admin",
      status: "active",
      usage_count: 8,
      created_at: "2026-02-01T12:00:00.000Z",
      expires_at: "2026-08-31T23:59:59.000Z",
      last_used_at: "2026-03-25T20:20:00.000Z",
      updated_at: "2026-03-25T20:20:00.000Z"
    }
  ],
  manual_token_form: {
    label_placeholder: "e.g. Internal ETL token",
    expiry_presets: ["End of semester", "30 days", "Custom date"],
    key_type_options: ["standard", "admin"]
  },
  server: {
    ec2_instance_id: "i-0abc123def4567890",
    state: "running",
    instance_type: "t3.large",
    public_ip: "34.130.220.91",
    uptime_human: "2d 14h 21m",
    hackutd_mode_enabled: true,
    hackutd_actions: [
      "stop instance",
      "set instance type to t3.large",
      "start instance"
    ],
    normal_mode_actions: [
      "stop instance",
      "set instance type to t3.micro",
      "start instance"
    ]
  },
  scraperPipeline: {
    last_run_at: "2026-03-25T18:40:00.000Z",
    status: "success",
    courses_uploaded: 1342,
    merge_match_rate: "96.2%",
    run_steps: ["scrape", "merge", "upload-dev"],
    latest_errors: []
  },
  analytics: {
    requests_per_day: [
      { day: "Mon", count: 1200, errors: 12 },
      { day: "Tue", count: 1420, errors: 14 },
      { day: "Wed", count: 1380, errors: 9 },
      { day: "Thu", count: 1690, errors: 22 },
      { day: "Fri", count: 1580, errors: 11 },
      { day: "Sat", count: 830, errors: 6 },
      { day: "Sun", count: 760, errors: 5 }
    ],
    by_token: [
      { label: "Course recommendation microservice", requests: 610 },
      { label: "HackUTD team service", requests: 482 },
      { label: "Officer emergency admin key", requests: 122 }
    ],
    by_endpoint: [
      { endpoint: "/v1/courses/search", requests: 790 },
      { endpoint: "/v1/grades", requests: 410 },
      { endpoint: "/v1/professors", requests: 380 },
      { endpoint: "/v1/terms", requests: 134 }
    ],
    error_rate_percent: 1.4
  },
  scraperLogs: [
    {
      time: "2026-03-25T18:41:20.000Z",
      level: "info",
      message: "Pipeline completed. Uploaded 1342 courses to dev dataset."
    },
    {
      time: "2026-03-25T18:40:31.000Z",
      level: "info",
      message: "Merge stage complete. Match rate 96.2%."
    },
    {
      time: "2026-03-25T18:37:02.000Z",
      level: "info",
      message: "Scraper stage complete. 11 college files updated."
    }
  ],
  cronLogs: [
    {
      time: "2026-03-25T12:00:00.000Z",
      tokens_checked: 401,
      valid_count: 387,
      action: "no-op"
    },
    {
      time: "2026-03-24T12:00:00.000Z",
      tokens_checked: 389,
      valid_count: 380,
      action: "started"
    },
    {
      time: "2026-03-23T12:00:00.000Z",
      tokens_checked: 384,
      valid_count: 369,
      action: "stopped"
    }
  ]
};
