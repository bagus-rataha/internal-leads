// Role values come back from the API as raw enum strings (e.g.
// ADMIN_SALES) - this maps each to the human-readable label the mockup
// actually shows next to a user's name.
export const ROLE_LABELS: Record<string, string> = {
  SALES: 'Sales',
  LEADER: 'Leader',
  ADMIN_SALES: 'Admin Sales',
  SU: 'Super User',
}

export function roleLabel(role: string | undefined): string {
  return (role && ROLE_LABELS[role]) ?? role ?? ''
}
