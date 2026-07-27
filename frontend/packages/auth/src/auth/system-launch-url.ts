export type LaunchableSystem = {
  launchUrl?: string;
  authenticatedLaunchUrl?: string;
};

export type SystemLaunchRuntime = {
  apiBaseUrl: string;
  hasPortalSession: boolean;
};

function isAbsoluteHttpURL(value: string): boolean {
  try {
    const url = new URL(value);
    return ["http:", "https:"].includes(url.protocol) && !url.username && !url.password;
  } catch {
    return false;
  }
}

function resolveBackendPath(path: string, apiBaseUrl: string): string {
  const base = apiBaseUrl.trim().replace(/\/$/, "");
  if (!base) return path;
  if (isAbsoluteHttpURL(path)) return path;

  return `${base}${path.startsWith("/") ? path : `/${path}`}`;
}

export function resolveSystemLaunchUrl(
  system: LaunchableSystem,
  runtime: SystemLaunchRuntime,
): string {
  const launchUrl = system.launchUrl?.trim() ?? "";
  const authenticatedLaunchUrl = system.authenticatedLaunchUrl?.trim() ?? "";

  if (!runtime.hasPortalSession && authenticatedLaunchUrl) {
    return resolveBackendPath(authenticatedLaunchUrl, runtime.apiBaseUrl);
  }

  return launchUrl || resolveBackendPath(authenticatedLaunchUrl, runtime.apiBaseUrl);
}
