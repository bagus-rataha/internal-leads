// Index content for the shell while Dashboard/Lead pages don't exist yet.
import { useAuth } from '@/auth/AuthContext'

export function WelcomeHome() {
  const { user } = useAuth()

  return (
    <div className="p-6">
      <h1 className="font-display font-extrabold text-xl">Selamat datang, {user?.name}</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Pilih menu di sidebar untuk mulai bekerja.
      </p>
    </div>
  )
}
