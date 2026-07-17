// Presentational top bar. Search is a visual placeholder (disabled) — wired
// up when the Lead list view lands.
import { Menu, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface TopbarProps {
  onMenuClick: () => void
}

export function Topbar({ onMenuClick }: TopbarProps) {
  return (
    <header className="flex h-[60px] shrink-0 items-center gap-3 border-b bg-background px-4">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        aria-label="Buka menu"
        className="lg:hidden"
        onClick={onMenuClick}
      >
        <Menu className="size-5" />
      </Button>

      <div className="relative hidden w-full max-w-md sm:block">
        <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Cari kode lead atau nama perusahaan…"
          disabled
          className="pl-8"
        />
      </div>
    </header>
  )
}
