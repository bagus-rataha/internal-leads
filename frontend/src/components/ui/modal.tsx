import type { ReactNode } from 'react'

export function Modal({
  open,
  onClose,
  children,
  maxWidth = 460,
}: {
  open: boolean
  onClose: () => void
  children: ReactNode
  maxWidth?: number
}) {
  if (!open) return null
  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-[rgba(15,23,42,0.5)] backdrop-blur-[2px]"
      onClick={onClose}
    >
      <div
        className="w-full rounded-[18px] bg-white p-[26px] shadow-[0_24px_48px_-12px_rgba(15,23,42,0.35)]"
        style={{ maxWidth }}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  )
}
