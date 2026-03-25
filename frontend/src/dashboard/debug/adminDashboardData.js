export const debugAdminDashboardData = {
  user: {
    email: "debug-admin@local.test"
  },
  requests: [
    {
      uid: "debug-user-001",
      email: "student1@utdallas.edu",
      requested_at: "2026-03-25T18:30:00.000Z",
      request_type: "new",
      status: "pending"
    },
    {
      uid: "debug-user-002",
      email: "student2@utdallas.edu",
      requested_at: "2026-03-25T17:15:00.000Z",
      request_type: "regenerate",
      status: "pending"
    }
  ],
  keys: [
    {
      uid: "debug-user-010",
      email: "activeuser@utdallas.edu",
      masked_key: "acm_****_debug",
      status: "active",
      usage_count: 42,
      updated_at: "2026-03-25T19:05:00.000Z"
    },
    {
      uid: "debug-user-011",
      email: "inactiveuser@utdallas.edu",
      masked_key: "acm_****_old",
      status: "inactive",
      usage_count: 3,
      updated_at: "2026-03-24T08:10:00.000Z"
    }
  ]
};
