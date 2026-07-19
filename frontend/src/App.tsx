import { Routes, Route } from 'react-router-dom'
import { RouteGuard } from '@/auth/RouteGuard'
import { AdminRouteGuard } from '@/auth/AdminRouteGuard'
import { LoginPage } from '@/features/auth/LoginPage'
import { Shell } from '@/components/layout/Shell'
import DashboardPage from '@/features/dashboard/DashboardPage'
import LeadListPage from '@/features/lead/LeadListPage'
import LeadDetailPage from '@/features/lead/LeadDetailPage'
import LeadFormPage from '@/features/lead/LeadFormPage'
import SalesTeamPage from '@/features/team/SalesTeamPage'
import LeadSourcePage from '@/features/reference/LeadSourcePage'
import TypeLayananPage from '@/features/reference/TypeLayananPage'
import UserPage from '@/features/user/UserPage'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RouteGuard><Shell /></RouteGuard>}>
        <Route index element={<DashboardPage />} />
        <Route path="leads" element={<LeadListPage />} />
        <Route path="leads/new" element={<LeadFormPage />} />
        <Route path="leads/:code" element={<LeadDetailPage />} />
        <Route path="leads/:code/edit" element={<LeadFormPage />} />
        <Route path="data/sales-team" element={<AdminRouteGuard><SalesTeamPage /></AdminRouteGuard>} />
        <Route path="data/lead-sources" element={<AdminRouteGuard><LeadSourcePage /></AdminRouteGuard>} />
        <Route path="data/service-types" element={<AdminRouteGuard><TypeLayananPage /></AdminRouteGuard>} />
        <Route path="settings/users" element={<AdminRouteGuard><UserPage /></AdminRouteGuard>} />
      </Route>
    </Routes>
  )
}

export default App
