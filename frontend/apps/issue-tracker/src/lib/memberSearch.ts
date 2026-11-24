import type { KeycloakGroupMember } from "../types";

export function formatMemberLabel(
  item: KeycloakGroupMember | null | undefined,
): string {
  if (!item) return "";
  const name = [item.firstName, item.lastName].filter(Boolean).join(" ").trim();
  return name
    ? `${name} (${item.email || item.username})`
    : item.email || item.username || "";
}

/**
 * Every whitespace-separated token of the query must appear in the member's
 * first name, last name, username or email. Mirrors the backend matcher
 * (matchesKeycloakUserSearch) so client-side filtering never hides a member
 * the server returned, and word order / extra spaces don't matter.
 */
export function memberMatchesSearch(
  item: KeycloakGroupMember,
  query: string | null | undefined,
): boolean {
  const tokens = (query ?? "").toLowerCase().split(/\s+/).filter(Boolean);
  if (tokens.length === 0) return true;
  const haystack = [item.firstName, item.lastName, item.username, item.email]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
  return tokens.every((token) => haystack.includes(token));
}
