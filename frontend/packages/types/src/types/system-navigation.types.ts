export type NavigationLaunchMode = "internal" | "new_tab" | "same_tab";

export interface SystemNavigationItem {
  id: string;
  label: string;
  path?: string;
  permission?: string;
  requiredPermissions?: string[];
  requiredAnyPermissions?: string[];
  order?: number;
  icon?: string;
  description?: string;
  displayInLauncher?: boolean;
  displayInSideNav?: boolean;
  launchMode?: NavigationLaunchMode;
  children?: SystemNavigationItem[];
}
