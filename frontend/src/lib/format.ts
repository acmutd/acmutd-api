export function formatDateTime(iso: string | null): string {
  if (!iso) {
    return "Never";
  }
  return new Date(iso).toLocaleString();
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString();
}

export function formatPercent(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

export function maskKey(value: string): string {
  if (value.length <= 10) {
    return "•".repeat(value.length);
  }
  return `${value.slice(0, 4)}${"•".repeat(Math.max(value.length - 8, 6))}${value.slice(-4)}`;
}

export function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${h}h ${m}m`;
}
