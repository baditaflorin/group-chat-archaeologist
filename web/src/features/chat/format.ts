export function formatDate(value: string) {
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric', year: 'numeric' }).format(
    new Date(value)
  );
}

export function formatMonth(value: string) {
  return new Intl.DateTimeFormat('en', { month: 'short', year: 'numeric' }).format(new Date(value));
}

export function daysAgo(value: string) {
  const then = new Date(value).getTime();
  const now = Date.now();
  const days = Math.max(0, Math.round((now - then) / 86_400_000));
  if (days === 0) {
    return 'today';
  }
  if (days === 1) {
    return '1 day ago';
  }
  return `${days} days ago`;
}
