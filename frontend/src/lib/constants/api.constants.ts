const API_ROOT = import.meta.env.VITE_API_BASE_URL + "/api";
const API_VERSION = "v1";

const API_BASE = `${API_ROOT}/${API_VERSION}`;

export const API = {
  root: API_ROOT,
  version: API_VERSION,
  base: API_BASE,

  // --------------------------------------------------
  // Auth
  // --------------------------------------------------
  auth: {
    base: `${API_BASE}/auth`,
    login: () => `${API_BASE}/auth/login`,
    callback: () => `${API_BASE}/auth/callback`,
    refresh: () => `${API_BASE}/auth/refresh`,
    logout: () => `${API_BASE}/auth/logout`,
    me: () => `${API_BASE}/auth/me`,
  },

  // --------------------------------------------------
  // Clients
  // --------------------------------------------------
  clients: {
    base: `${API_BASE}/clients`,
    list: () => `${API_BASE}/clients`,
    byId: (id: string) => `${API_BASE}/clients/${id}`,
    create: () => `${API_BASE}/clients`,
    delete: (id: string) => `${API_BASE}/clients/${id}`,
  },

  // --------------------------------------------------
  // Users
  // --------------------------------------------------
  users: {
    base: `${API_BASE}/users`,
    list: () => `${API_BASE}/users`,
    byId: (id: string) => `${API_BASE}/users/${id}`,
    create: () => `${API_BASE}/users`,
    delete: (id: string) => `${API_BASE}/users/${id}`,
  },

  // --------------------------------------------------
  // Admin
  // --------------------------------------------------
  admin: {
    base: `${API_BASE}/admin`,

    users: {
      list: () => `${API_BASE}/admin/users`,
      byId: (id: string) => `${API_BASE}/admin/users/${id}`,
      create: () => `${API_BASE}/admin/users`,
      delete: (id: string) => `${API_BASE}/admin/users/${id}`,

      import: {
        preview: () => `${API_BASE}/admin/users/import/preview`,
        execute: () => `${API_BASE}/admin/users/import/execute`,
        job: (jobId: string) => `${API_BASE}/admin/users/import/${jobId}`,
        errorsCsv: (jobId: string) =>
          `${API_BASE}/admin/users/import/${jobId}/errors.csv`,
        templateCsv: () => `${API_BASE}/admin/users/import/template.csv`,
      },
    },

    metrics: {
      overview: () => `${API_BASE}/admin/metrics/overview`,

      system: {
        countUsers: () => `${API_BASE}/admin/metrics/system/count-users`,
        countDisabledUsers: () =>
          `${API_BASE}/admin/metrics/system/count-disabled-users`,
        activeToday: () => `${API_BASE}/admin/metrics/system/active-today`,
        activeThisWeek: () =>
          `${API_BASE}/admin/metrics/system/active-this-week`,
        loginTrend: () => `${API_BASE}/admin/metrics/system/login-trend`,
        loginTrendRange: () =>
          `${API_BASE}/admin/metrics/system/login-trend-range`,
      },

      security: {
        failedLogins: () => `${API_BASE}/admin/metrics/security/failed-logins`,
        failedLoginsRange: () =>
          `${API_BASE}/admin/metrics/security/failed-logins-range`,
        suspiciousLogins: () =>
          `${API_BASE}/admin/metrics/security/suspicious-logins`,
      },

      clients: {
        count: () => `${API_BASE}/admin/metrics/clients/count`,
        mostAccessed: () => `${API_BASE}/admin/metrics/clients/most-accessed`,
        loginCount: () => `${API_BASE}/admin/metrics/clients/login-count`,
        activeToday: () => `${API_BASE}/admin/metrics/clients/active-today`,
      },

      users: {
        newRange: () => `${API_BASE}/admin/metrics/users/new-range`,
        newTrend: () => `${API_BASE}/admin/metrics/users/new-trend`,
        neverLoggedIn: () => `${API_BASE}/admin/metrics/users/never-logged-in`,
        lastLogin: (userId: string) =>
          `${API_BASE}/admin/metrics/users/last-login/${userId}`,
        clientUsage: (userId: string) =>
          `${API_BASE}/admin/metrics/users/client-usage/${userId}`,
      },
    },

    audit: {
      base: `${API_BASE}/admin/audit-logs`,
      list: () => `${API_BASE}/admin/audit-logs`,
      actions: () => `${API_BASE}/admin/audit-logs/actions`,
      byId: (id: string) => `${API_BASE}/admin/audit-logs/${id}`,

      metrics: {
        overview: () => `${API_BASE}/admin/audit-logs/metrics/overview`,
        failedLoginsByDay: () =>
          `${API_BASE}/admin/audit-logs/metrics/failed-logins-by-day`,
        topFailureIps: () =>
          `${API_BASE}/admin/audit-logs/metrics/top-failure-ips`,
      },

      export: () => `${API_BASE}/admin/audit-logs/export`,
    },

    notifications: {
      base: `${API_BASE}/admin/notifications`,
      notify: () => `${API_BASE}/admin/notifications`,
      list: () => `${API_BASE}/admin/notifications`,
      byId: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      markAsRead: (id: string) => `${API_BASE}/admin/notifications/${id}/read`,
      delete: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      count: () => `${API_BASE}/admin/notifications/count`,
      countUnread: () => `${API_BASE}/admin/notifications/count/unread`,
      deleteOld: () => `${API_BASE}/admin/notifications/cleanup`,
    },
  },

  // --------------------------------------------------
  // Health (not versioned)
  // --------------------------------------------------
  health: () => `/health`,
} as const;
