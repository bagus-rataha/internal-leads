// Single-slot toast: a tiny module-level pub/sub, not a context provider -
// there's exactly one <Toast/> mount point (Shell.tsx) and at most one
// visible message at a time across this whole app.
import { useEffect, useState } from 'react'

interface ToastState {
  message: string
  key: number
}

let listener: ((state: ToastState) => void) | null = null
let counter = 0

export function showToast(message: string) {
  counter += 1
  listener?.({ message, key: counter })
}

export function useToastState(): ToastState | null {
  const [state, setState] = useState<ToastState | null>(null)

  useEffect(() => {
    listener = setState
    return () => {
      listener = null
    }
  }, [])

  return state
}
