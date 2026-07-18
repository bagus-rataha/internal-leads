// Generic list + create/edit modal for the 3 structurally identical master
// data entities (Sales Team, Sumber Lead, Type Layanan - all just `name` +
// `is_active`). Each entity's page wires its own query/mutation hooks into
// this component instead of three near-duplicate table+modal files.
import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Modal } from '@/components/ui/modal'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { showToast } from '@/hooks/useToast'

interface SimpleEntity {
  id: string
  name: string
  is_active: boolean
}

interface SimpleEntityListProps<T extends SimpleEntity> {
  title: string
  addLabel: string
  items: T[] | undefined
  isLoading: boolean
  onCreate: (name: string) => Promise<unknown>
  onUpdate: (id: string, input: { name?: string; is_active?: boolean }) => Promise<unknown>
  isSaving: boolean
}

export function SimpleEntityList<T extends SimpleEntity>({
  title,
  addLabel,
  items,
  isLoading,
  onCreate,
  onUpdate,
  isSaving,
}: SimpleEntityListProps<T>) {
  const [modal, setModal] = useState<{ mode: 'create' } | { mode: 'edit'; item: T } | null>(null)
  const [name, setName] = useState('')
  const [isActive, setIsActive] = useState(true)

  function openCreate() {
    setName('')
    setIsActive(true)
    setModal({ mode: 'create' })
  }

  function openEdit(item: T) {
    setName(item.name)
    setIsActive(item.is_active)
    setModal({ mode: 'edit', item })
  }

  async function handleSubmit() {
    if (!name.trim()) {
      showToast('Nama tidak boleh kosong')
      return
    }
    try {
      if (modal?.mode === 'create') {
        await onCreate(name.trim())
        showToast(`${title} ditambahkan`)
      } else if (modal?.mode === 'edit') {
        await onUpdate(modal.item.id, { name: name.trim(), is_active: isActive })
        showToast(`${title} diperbarui`)
      }
      setModal(null)
    } catch {
      showToast(`Gagal menyimpan ${title.toLowerCase()}. Coba lagi.`)
    }
  }

  return (
    <div className="p-4 lg:p-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">{title}</h1>
        <Button onClick={openCreate}>{addLabel}</Button>
      </div>

      <Card className="overflow-hidden">
        {isLoading ? (
          <p className="p-4 text-sm text-muted-foreground">Memuat...</p>
        ) : !items?.length ? (
          <p className="p-4 text-sm text-muted-foreground">Belum ada data.</p>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-[#F7F9FC] text-left text-[10.5px] font-bold uppercase text-[#94A3B8]">
              <tr>
                <th className="px-4 py-2.5">Nama</th>
                <th className="px-4 py-2.5">Status</th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-[#F1F5F9]">
              {items.map((item) => (
                <tr key={item.id}>
                  <td className="px-4 py-2.5">{item.name}</td>
                  <td className="px-4 py-2.5">
                    <span
                      className={
                        item.is_active
                          ? 'rounded-full bg-[#DCFCE7] px-2 py-0.5 text-xs font-medium text-[#166534]'
                          : 'rounded-full bg-[#F1F5F9] px-2 py-0.5 text-xs font-medium text-[#64748B]'
                      }
                    >
                      {item.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-right">
                    <Button variant="outline" size="sm" onClick={() => openEdit(item)}>
                      Edit
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Modal open={modal !== null} onClose={() => setModal(null)}>
        <h2 className="mb-4 text-lg font-semibold">
          {modal?.mode === 'edit' ? `Edit ${title}` : addLabel}
        </h2>
        <div className="flex flex-col gap-3">
          <div>
            <Label htmlFor="entity-name">Nama</Label>
            <Input id="entity-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          {modal?.mode === 'edit' && (
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
              />
              Aktif
            </label>
          )}
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" onClick={() => setModal(null)} disabled={isSaving}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isSaving}>
            Simpan
          </Button>
        </div>
      </Modal>
    </div>
  )
}
