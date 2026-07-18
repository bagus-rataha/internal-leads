import { useState } from 'react'
import { MoreHorizontal } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Popover, PopoverClose, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { showToast } from '@/hooks/useToast'
import { useAuth } from '@/auth/AuthContext'
import { useTeams } from '@/features/team/queries'
import { useUsers, useCreateUser, useUpdateUser, useResetPassword, useDeactivateUser } from './queries'
import { ActiveLeadsError, EmailTakenError, type UserResponse } from './api'

const ROLES = ['SALES', 'LEADER', 'ADMIN_SALES', 'SU'] as const
type Role = (typeof ROLES)[number]

function needsTeam(role: string) {
  return role === 'SALES' || role === 'LEADER'
}

function UserActionsMenu({
  user,
  onEdit,
  onResetPassword,
  onDeactivate,
  onReactivate,
  reactivatePending,
}: {
  user: UserResponse
  onEdit: () => void
  onResetPassword: () => void
  onDeactivate: () => void
  onReactivate: () => void
  reactivatePending: boolean
}) {
  return (
    <Popover>
      <PopoverTrigger
        aria-label={`Aksi untuk ${user.name}`}
        className="flex size-7 items-center justify-center rounded-lg border border-input bg-background hover:bg-muted"
      >
        <MoreHorizontal className="size-4" />
      </PopoverTrigger>
      <PopoverContent align="end" className="w-44 gap-0.5 p-1">
        <PopoverClose
          className="w-full rounded-md px-2.5 py-1.5 text-left text-sm hover:bg-muted"
          onClick={onEdit}
        >
          Edit
        </PopoverClose>
        <PopoverClose
          className="w-full rounded-md px-2.5 py-1.5 text-left text-sm hover:bg-muted"
          onClick={onResetPassword}
        >
          Reset Password
        </PopoverClose>
        {user.is_active ? (
          <PopoverClose
            className="w-full rounded-md px-2.5 py-1.5 text-left text-sm text-destructive hover:bg-destructive/10"
            onClick={onDeactivate}
          >
            Nonaktifkan
          </PopoverClose>
        ) : (
          <PopoverClose
            className="w-full rounded-md px-2.5 py-1.5 text-left text-sm hover:bg-muted disabled:pointer-events-none disabled:opacity-50"
            disabled={reactivatePending}
            onClick={onReactivate}
          >
            Aktifkan
          </PopoverClose>
        )}
      </PopoverContent>
    </Popover>
  )
}

export default function UserPage() {
  const { user: currentUser } = useAuth()
  const [roleFilter, setRoleFilter] = useState('')
  const { data: users, isLoading } = useUsers(roleFilter || undefined)
  // Unfiltered, independent of roleFilter - the reassign dropdown must offer
  // every active user regardless of what role the page happens to be
  // filtered to (the backend accepts any valid target id, not just SALES).
  const { data: allUsers } = useUsers()
  const { data: teams } = useTeams()
  const createUser = useCreateUser()
  const updateUser = useUpdateUser()
  const resetPassword = useResetPassword()
  const deactivateUser = useDeactivateUser()

  const [modal, setModal] = useState<{ mode: 'create' } | { mode: 'edit'; user: UserResponse } | null>(null)
  const [resetTarget, setResetTarget] = useState<UserResponse | null>(null)
  const [confirmDeactivateTarget, setConfirmDeactivateTarget] = useState<UserResponse | null>(null)
  const [emailTakenError, setEmailTakenError] = useState<EmailTakenError | null>(null)
  const [deactivateTarget, setDeactivateTarget] = useState<UserResponse | null>(null)
  const [reassignCount, setReassignCount] = useState<number | null>(null)
  const [reassignToId, setReassignToId] = useState('')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('SALES')
  const [teamId, setTeamId] = useState('')
  const [newPassword, setNewPassword] = useState('')

  function openCreate() {
    setName('')
    setEmail('')
    setPassword('')
    setRole('SALES')
    setTeamId('')
    setModal({ mode: 'create' })
  }

  function openEdit(u: UserResponse) {
    setName(u.name ?? '')
    setRole(u.role as Role)
    setTeamId(u.team_id ?? '')
    setModal({ mode: 'edit', user: u })
  }

  const saving = createUser.isPending || updateUser.isPending

  async function handleSubmit() {
    if (!name.trim()) {
      showToast('Nama tidak boleh kosong')
      return
    }
    try {
      if (modal?.mode === 'create') {
        if (!email.trim() || !password) {
          showToast('Email dan password wajib diisi')
          return
        }
        await createUser.mutateAsync({
          name: name.trim(),
          email: email.trim(),
          password,
          role,
          ...(needsTeam(role) && teamId ? { team_id: teamId } : {}),
        })
        showToast('User ditambahkan')
      } else if (modal?.mode === 'edit') {
        await updateUser.mutateAsync({
          id: modal.user.id ?? '',
          input: {
            name: name.trim(),
            role,
            // Omitted (not null) to satisfy the generated type - the backend
            // treats an absent team_id identically to an explicit null for
            // this pointer field (see user_service.go's updateUser), so this
            // is behavior-preserving.
            team_id: needsTeam(role) && teamId ? teamId : undefined,
          },
        })
        showToast('User diperbarui')
      }
      setModal(null)
    } catch (err) {
      if (err instanceof EmailTakenError) {
        if (err.existingUserId && !err.existingUserActive) {
          setEmailTakenError(err)
        } else if (err.existingUserActive) {
          showToast(`Email sudah terdaftar atas nama ${err.existingUserName}`)
        } else {
          showToast('Email sudah terdaftar')
        }
      } else {
        showToast(err instanceof Error ? err.message : 'Gagal menyimpan user. Coba lagi.')
      }
    }
  }

  async function handleResetPassword() {
    if (!resetTarget) return
    if (newPassword.length < 6) {
      showToast('Password minimal 6 karakter')
      return
    }
    try {
      await resetPassword.mutateAsync({
        id: resetTarget.id ?? '',
        input: { new_password: newPassword },
      })
      showToast('Password direset')
      setResetTarget(null)
      setNewPassword('')
    } catch {
      showToast('Gagal mereset password. Coba lagi.')
    }
  }

  async function handleDeactivate(target: UserResponse) {
    try {
      await deactivateUser.mutateAsync({ id: target.id ?? '', input: {} })
      showToast('User dinonaktifkan')
    } catch (err) {
      if (err instanceof ActiveLeadsError) {
        setDeactivateTarget(target)
        setReassignCount(err.activeLeadCount)
        setReassignToId('')
      } else {
        showToast('Gagal menonaktifkan user. Coba lagi.')
      }
    }
  }

  async function handleReactivate(id: string, name?: string) {
    try {
      await updateUser.mutateAsync({ id, input: { is_active: true } })
      showToast(name ? `${name} diaktifkan kembali` : 'User diaktifkan kembali')
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Gagal mengaktifkan user. Coba lagi.')
    }
  }

  async function handleReassignAndDeactivate() {
    if (!deactivateTarget || !reassignToId) {
      showToast('Pilih sales pengganti')
      return
    }
    try {
      await deactivateUser.mutateAsync({
        id: deactivateTarget.id ?? '',
        input: { reassign_to_user_id: reassignToId },
      })
      showToast('Lead dipindahkan, user dinonaktifkan')
      setDeactivateTarget(null)
      setReassignCount(null)
    } catch {
      showToast('Gagal memindahkan lead. Coba lagi.')
    }
  }

  return (
    <div className="p-4 lg:p-6">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-semibold">User</h1>
        <div className="flex items-center gap-2">
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="h-8 rounded-lg border border-input px-2 text-sm"
          >
            <option value="">Semua Role</option>
            {ROLES.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
          <Button onClick={openCreate}>Tambah User</Button>
        </div>
      </div>

      <Card className="overflow-hidden">
        {isLoading ? (
          <p className="p-4 text-sm text-muted-foreground">Memuat...</p>
        ) : !users?.length ? (
          <p className="p-4 text-sm text-muted-foreground">Belum ada data.</p>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-[#F7F9FC] text-left text-[10.5px] font-bold uppercase text-[#94A3B8]">
              <tr>
                <th className="px-4 py-2.5">Nama</th>
                <th className="px-4 py-2.5">Email</th>
                <th className="px-4 py-2.5">Role</th>
                <th className="px-4 py-2.5">Tim</th>
                <th className="px-4 py-2.5">Status</th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-[#F1F5F9]">
              {users.map((u) => {
                // Backend rejects Edit/Reset Password/Deactivate for ADMIN_SALES acting on an SU
                // target - hiding here is convenience only, the real boundary is server-side.
                const isForbiddenTarget = currentUser?.role === 'ADMIN_SALES' && u.role === 'SU'
                return (
                <tr key={u.id}>
                  <td className="px-4 py-2.5">{u.name}</td>
                  <td className="px-4 py-2.5">{u.email}</td>
                  <td className="px-4 py-2.5">{u.role}</td>
                  <td className="px-4 py-2.5">
                    {teams?.find((t) => t.id === u.team_id)?.name ?? '—'}
                  </td>
                  <td className="px-4 py-2.5">
                    <span
                      className={
                        u.is_active
                          ? 'rounded-full bg-[#DCFCE7] px-2 py-0.5 text-xs font-medium text-[#166534]'
                          : 'rounded-full bg-[#F1F5F9] px-2 py-0.5 text-xs font-medium text-[#64748B]'
                      }
                    >
                      {u.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-right">
                    <div className="flex justify-end gap-2">
                      {isForbiddenTarget ? (
                        <span className="text-xs text-muted-foreground">Tidak dapat dikelola</span>
                      ) : (
                        <UserActionsMenu
                          user={u}
                          onEdit={() => openEdit(u)}
                          onResetPassword={() => setResetTarget(u)}
                          onDeactivate={() => setConfirmDeactivateTarget(u)}
                          onReactivate={() => handleReactivate(u.id ?? '', u.name)}
                          reactivatePending={updateUser.isPending}
                        />
                      )}
                    </div>
                  </td>
                </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </Card>

      <Modal open={modal !== null} onClose={() => setModal(null)}>
        <h2 className="mb-4 text-lg font-semibold">
          {modal?.mode === 'edit' ? 'Edit User' : 'Tambah User'}
        </h2>
        <div className="flex flex-col gap-3">
          <div>
            <Label htmlFor="user-name">Nama</Label>
            <Input id="user-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          {modal?.mode === 'create' && (
            <>
              <div>
                <Label htmlFor="user-email">Email</Label>
                <Input
                  id="user-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              <div>
                <Label htmlFor="user-password">Password</Label>
                <Input
                  id="user-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            </>
          )}
          <div>
            <Label htmlFor="user-role">Role</Label>
            <select
              id="user-role"
              value={role}
              onChange={(e) => setRole(e.target.value as Role)}
              className="h-8 w-full rounded-lg border border-input px-2 text-sm"
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </div>
          {needsTeam(role) && (
            <div>
              <Label htmlFor="user-team">Tim</Label>
              <select
                id="user-team"
                value={teamId}
                onChange={(e) => setTeamId(e.target.value)}
                className="h-8 w-full rounded-lg border border-input px-2 text-sm"
              >
                <option value="">Pilih tim</option>
                {teams?.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" onClick={() => setModal(null)} disabled={saving}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={saving}>
            Simpan
          </Button>
        </div>
      </Modal>

      <Modal open={resetTarget !== null} onClose={() => setResetTarget(null)} maxWidth={380}>
        <h2 className="mb-4 text-lg font-semibold">Reset Password — {resetTarget?.name}</h2>
        <Label htmlFor="new-password">Password Baru</Label>
        <Input
          id="new-password"
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
        />
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" onClick={() => setResetTarget(null)} disabled={resetPassword.isPending}>
            Batal
          </Button>
          <Button onClick={handleResetPassword} disabled={resetPassword.isPending}>
            Reset
          </Button>
        </div>
      </Modal>

      <Modal
        open={confirmDeactivateTarget !== null}
        onClose={() => setConfirmDeactivateTarget(null)}
        maxWidth={380}
      >
        <h2 className="mb-2 text-lg font-semibold text-destructive">Nonaktifkan User</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          {confirmDeactivateTarget?.name} tidak akan bisa login setelah dinonaktifkan. Tindakan ini
          bisa dibatalkan lewat tombol Aktifkan.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setConfirmDeactivateTarget(null)}>
            Batal
          </Button>
          <Button
            variant="destructive"
            onClick={() => {
              const target = confirmDeactivateTarget
              setConfirmDeactivateTarget(null)
              if (target) handleDeactivate(target)
            }}
          >
            Ya, Nonaktifkan
          </Button>
        </div>
      </Modal>

      <Modal open={emailTakenError !== null} onClose={() => setEmailTakenError(null)} maxWidth={380}>
        <h2 className="mb-2 text-lg font-semibold">Email Sudah Terdaftar</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          Email ini sudah dipakai oleh {emailTakenError?.existingUserName}, yang saat ini nonaktif.
          Aktifkan kembali akun tersebut daripada membuat user baru?
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setEmailTakenError(null)}>
            Batal
          </Button>
          <Button
            onClick={async () => {
              if (!emailTakenError) return
              await handleReactivate(emailTakenError.existingUserId, emailTakenError.existingUserName)
              setEmailTakenError(null)
              setModal(null)
            }}
          >
            Aktifkan {emailTakenError?.existingUserName}
          </Button>
        </div>
      </Modal>

      <Modal
        open={deactivateTarget !== null}
        onClose={() => {
          setDeactivateTarget(null)
          setReassignCount(null)
        }}
        maxWidth={420}
      >
        <h2 className="mb-2 text-lg font-semibold">Pindahkan Lead Aktif</h2>
        <p className="mb-4 text-sm text-muted-foreground">
          {deactivateTarget?.name} masih punya {reassignCount} lead aktif. Pilih sales pengganti
          sebelum menonaktifkan.
        </p>
        <Label htmlFor="reassign-to">Pindahkan ke</Label>
        <select
          id="reassign-to"
          value={reassignToId}
          onChange={(e) => setReassignToId(e.target.value)}
          className="h-8 w-full rounded-lg border border-input px-2 text-sm"
        >
          <option value="">Pilih user</option>
          {allUsers
            ?.filter((u) => u.id !== deactivateTarget?.id && u.is_active)
            .map((u) => (
              <option key={u.id} value={u.id}>
                {u.name} ({u.role})
              </option>
            ))}
        </select>
        <div className="mt-5 flex justify-end gap-2">
          <Button
            variant="outline"
            onClick={() => {
              setDeactivateTarget(null)
              setReassignCount(null)
            }}
            disabled={deactivateUser.isPending}
          >
            Batal
          </Button>
          <Button onClick={handleReassignAndDeactivate} disabled={deactivateUser.isPending}>
            Pindahkan &amp; Nonaktifkan
          </Button>
        </div>
      </Modal>
    </div>
  )
}
