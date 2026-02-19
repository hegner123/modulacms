import { useQuery } from '@tanstack/react-query'
import { sessionSdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

export function useMe() {
  return useQuery({
    queryKey: queryKeys.auth.me(),
    queryFn: () => sessionSdk.auth.me(),
    retry: false,
  })
}
