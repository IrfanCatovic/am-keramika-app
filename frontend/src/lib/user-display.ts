/** Prefer full name for people-facing labels; username is login identity only. */
export function userDisplayName(user: {
  fullName?: string | null;
  username?: string | null;
} | null | undefined): string {
  if (!user) {
    return "—";
  }
  const fullName = user.fullName?.trim();
  if (fullName) {
    return fullName;
  }
  const username = user.username?.trim();
  if (username) {
    return username;
  }
  return "—";
}
