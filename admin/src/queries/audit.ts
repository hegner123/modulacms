import { useQuery } from '@tanstack/react-query'

export function useAuditLog() {
  return useQuery({
    queryKey: [] as const,
    queryFn: () => Promise.resolve(null),
    enabled: false,
  })
}
