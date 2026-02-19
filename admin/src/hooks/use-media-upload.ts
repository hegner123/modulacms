import { useState, useCallback } from 'react'
import type { Media } from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'

type UseMediaUploadReturn = {
  upload: (file: File, path?: string) => Promise<Media>
  isUploading: boolean
  progress: number
  error: Error | null
  reset: () => void
}

export function useMediaUpload(): UseMediaUploadReturn {
  const [isUploading, setIsUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState<Error | null>(null)

  const reset = useCallback(() => {
    setIsUploading(false)
    setProgress(0)
    setError(null)
  }, [])

  const upload = useCallback(async (file: File, path?: string): Promise<Media> => {
    setIsUploading(true)
    setProgress(0)
    setError(null)

    try {
      setProgress(50)
      const opts = path ? { path } : undefined
      const media = await sdk.mediaUpload.upload(file, opts)
      setProgress(100)
      return media
    } catch (err) {
      const uploadError = err instanceof Error ? err : new Error('Upload failed')
      setError(uploadError)
      throw uploadError
    } finally {
      setIsUploading(false)
    }
  }, [])

  return { upload, isUploading, progress, error, reset }
}
