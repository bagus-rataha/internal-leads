import { Routes, Route } from 'react-router-dom'
import { RouteGuard } from '@/auth/RouteGuard'
import { LoginPage } from '@/features/auth/LoginPage'
import { Shell } from '@/components/layout/Shell'
import { WelcomeHome } from '@/components/layout/WelcomeHome'
import LeadListPage from '@/features/lead/LeadListPage'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RouteGuard><Shell /></RouteGuard>}>
        <Route index element={<WelcomeHome />} />
        <Route path="leads" element={<LeadListPage />} />
      </Route>
    </Routes>
  )
}

export default App
