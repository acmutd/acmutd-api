export const debugUserDashboardData = {
  user: {
    uid: "debug-user-001",
    name: "Alex Morgan",
    email: "alex.morgan@utdallas.edu",
    photo_url: "https://i.pravatar.cc/96?img=12",
    auth_provider: "google.com",
    domain_restricted: true
  },
  auth: {
    oauth_enabled: true,
    allowed_domain: "utdallas.edu",
    firebase_token_status: "valid"
  },
  dashboard: {
    status: "active",
    mode: "manual-approval",
    semester_expiry_policy: "current-semester",
    hackutd_mode: {
      enabled: true,
      auto_approve: true,
      ends_at: "2026-11-16T05:00:00.000Z",
      forced_expiry_at: "2026-11-17T17:00:00.000Z"
    },
    request_form: {
      default_key_type: "standard",
      name_placeholder: "e.g. ACM Discord bot",
      description_placeholder: "Short description for admins"
    },
    key_info: {
      masked_key: "acm_****_debug",
      full_key_shown_once: "acm_live_visible_once_x7f9v1",
      created_at: "2026-01-19T20:12:00.000Z",
      expires_at: "2026-05-14T23:59:59.000Z",
      last_rotated_at: "2026-02-28T16:04:00.000Z",
      key_type: "standard",
      rate_limit: 120,
      window_seconds: 60,
      usage_count: 743,
      is_admin_key: false,
      request_description: "Used for class project API integration"
    },
    usage: {
      total_requests: 743,
      last_used_at: "2026-03-25T17:02:10.000Z",
      endpoint_breakdown: [
        { endpoint: "/v1/courses/search", requests: 322 },
        { endpoint: "/v1/professors", requests: 216 },
        { endpoint: "/v1/grades", requests: 141 },
        { endpoint: "/v1/terms", requests: 64 }
      ],
      recent_requests: [
        { at: "2026-03-25T16:58:11.000Z", endpoint: "/v1/courses/search", status: 200 },
        { at: "2026-03-25T16:58:02.000Z", endpoint: "/v1/courses/search", status: 200 },
        { at: "2026-03-25T16:57:31.000Z", endpoint: "/v1/grades", status: 429 },
        { at: "2026-03-25T16:55:10.000Z", endpoint: "/v1/professors", status: 200 },
        { at: "2026-03-25T16:52:40.000Z", endpoint: "/v1/terms", status: 200 }
      ]
    },
    expiry_notifications: {
      email_warning_enabled: true,
      warning_windows_days: [14, 7, 1],
      next_warning_at: "2026-05-01T15:00:00.000Z"
    }
  },
  generatedKey: "acm_debug_generated_key_for_ui_preview",
  allowed_actions: {
    can_request_new_key: true,
    can_regenerate_key: true,
    can_revoke_key: true,
    max_keys_per_user: 1,
    admin_override_allows_multiple: true
  }
};
