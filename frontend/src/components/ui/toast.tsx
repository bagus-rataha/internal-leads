// Fixed-bottom, auto-dismissing (~2.6s, matching the mockup's own timing)
// toast. ponytail: plain show/hide via setTimeout, no fade transition -
// add one only if asked, the mockup's CSS keyframe fade isn't load-bearing.
import { useEffect, useState } from 'react'
import { useToastState } from '@/hooks/useToast'

export function Toast() {
  const state = useToastState()
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    if (!state) return
    setVisible(true)
    const timer = setTimeout(() => setVisible(false), 2600)
    return () => clearTimeout(timer)
  }, [state])

  if (!state || !visible) return null

  return (
    <div
      key={state.key}
      className="fixed bottom-[26px] left-1/2 z-[200] -translate-x-1/2 rounded-[11px] bg-[#0F172A] px-[18px] py-[11px] text-[13.5px] font-medium text-white shadow-[0_14px_32px_-8px_rgba(15,23,42,0.45)]"
    >
      {state.message}
    </div>
  )
}
