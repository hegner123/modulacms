import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { MediaID, UpdateMediaParams } from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

export function useMediaList() {
  return useQuery({
    queryKey: queryKeys.media.list(),
    queryFn: () => sdk.media.list(),
  })
}

export function useMedia(id: string) {
  return useQuery({
    queryKey: queryKeys.media.detail(id),
    queryFn: () => sdk.media.get(id as MediaID),
    enabled: !!id,
  })
}

export function useUpdateMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateMediaParams) => sdk.media.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.media.all })
    },
  })
}

export function useDeleteMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: MediaID) => sdk.media.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.media.all })
    },
  })
}
