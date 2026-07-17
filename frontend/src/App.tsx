import { Routes, Route } from 'react-router-dom'
import { useAuth } from '@/auth/AuthContext'
import { RouteGuard } from '@/auth/RouteGuard'
import { LoginPage } from '@/features/auth/LoginPage'
import { Button } from '@/components/ui/button'

// Placeholder only, proves RouteGuard protects a route. Real shell/navigation comes later.
function PlaceholderHome() {
  const { user, logout } = useAuth()
  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-4">
      <p>Logged in as {user?.name}</p>
      <Button onClick={() => logout()}>Logout</Button>
    </div>
  )
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<RouteGuard><PlaceholderHome /></RouteGuard>} />
    </Routes>
  )
}

export default App
