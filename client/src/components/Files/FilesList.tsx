import React, { useState } from 'react'
import { DataTable } from 'primereact/datatable'
import type { DataTableRowClickEvent } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { FileUpload } from 'primereact/fileupload'
import type { FileUploadHandlerEvent } from 'primereact/fileupload'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { Toast } from 'primereact/toast'
import type { FileEntry } from '../../services/filesApi'
import {
  useGetFilesQuery,
  useUploadFileMutation,
  useDeleteFileMutation,
  useCreateDirectoryMutation,
} from '../../services/filesApi'

interface FilesListProps {
  currentPath: string
  onNavigate: (path: string) => void
}

export function FilesList({ currentPath, onNavigate }: FilesListProps) {
  const toast = React.useRef<any>(null)
  const [uploadVisible, setUploadVisible] = useState(false)
  const [mkdirVisible, setMkdirVisible] = useState(false)
  const [newDirName, setNewDirName] = useState('')
  const [selectedFile, setSelectedFile] = useState<FileEntry | null>(null)

  const { data: files, isLoading, refetch } = useGetFilesQuery({ path: currentPath })

  const [uploadFile, { isLoading: isUploading }] = useUploadFileMutation()
  const [deleteFile, { isLoading: isDeleting }] = useDeleteFileMutation()
  const [createDirectory, { isLoading: isCreating }] = useCreateDirectoryMutation()

  const formatSize = (size: number) => {
    const units = ['B', 'KB', 'MB', 'GB']
    let i = 0
    let s = size
    while (s >= 1024 && i < units.length - 1) {
      s /= 1024
      i++
    }
    return `${s.toFixed(s >= 10 ? 0 : 2)} ${units[i]}`
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return '-'
    const d = new Date(dateStr)
    return d.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const getFileIcon = (file: FileEntry) => {
    if (file.type === 'folder') return 'pi pi-folder text-yellow-500 text-2xl'
    const ext = file.name.split('.').pop()?.toLowerCase()
    const icons: Record<string, string> = {
      pdf: 'pi pi-file-pdf text-red-500',
      doc: 'pi pi-file-word text-blue-500',
      docx: 'pi pi-file-word text-blue-500',
      xls: 'pi pi-file-excel text-green-500',
      xlsx: 'pi pi-file-excel text-green-500',
      jpg: 'pi pi-image text-purple-500',
      jpeg: 'pi pi-image text-purple-500',
      png: 'pi pi-image text-purple-500',
      gif: 'pi pi-image text-purple-500',
      zip: 'pi pi-box text-orange-500',
      rar: 'pi pi-box text-orange-500',
      txt: 'pi pi-file text-gray-500',
      sh: 'pi pi-cog text-gray-700',
      js: 'pi pi-file text-yellow-600',
      ts: 'pi pi-file text-blue-600',
      html: 'pi pi-code text-orange-600',
      css: 'pi pi-code text-pink-500',
      json: 'pi pi-code text-yellow-700',
      yml: 'pi pi-code text-gray-600',
      yaml: 'pi pi-code text-gray-600',
      sql: 'pi pi-database text-indigo-500',
      env: 'pi pi-lock text-red-600',
    }
    return icons[ext || ''] || 'pi pi-file text-gray-500'
  }

  const handleNavigate = (file: FileEntry) => {
    if (file.type === 'folder') {
      onNavigate(file.path)
    }
  }

  const handleUpload = async (event: FileUploadHandlerEvent) => {
    const files = event.files || []
    if (files.length === 0) return
    const file = files[0] as File
    try {
      await uploadFile({ file, path: currentPath }).unwrap()
      toast.current?.show({
        severity: 'success',
        summary: 'Успешно',
        detail: 'Файл загружен',
        life: 3000,
      })
      setUploadVisible(false)
      refetch()
    } catch (err: any) {
      toast.current?.show({
        severity: 'error',
        summary: 'Ошибка',
        detail: err.data?.error || err.error || 'Не удалось загрузить файл',
        life: 5000,
      })
    }
  }

  const handleDelete = async () => {
    if (!selectedFile) return
    try {
      await deleteFile(selectedFile.path).unwrap()
      toast.current?.show({
        severity: 'success',
        summary: 'Успешно',
        detail: 'Удалено',
        life: 3000,
      })
      setSelectedFile(null)
      refetch()
    } catch (err: any) {
      toast.current?.show({
        severity: 'error',
        summary: 'Ошибка',
        detail: err.data?.error || err.error || 'Не удалось удалить',
        life: 5000,
      })
    }
  }

  const handleCreateDirectory = async () => {
    if (!newDirName.trim()) return
    try {
      await createDirectory({ path: currentPath, name: newDirName.trim() }).unwrap()
      toast.current?.show({
        severity: 'success',
        summary: 'Успешно',
        detail: 'Папка создана',
        life: 3000,
      })
      setNewDirName('')
      setMkdirVisible(false)
      refetch()
    } catch (err: any) {
      toast.current?.show({
        severity: 'error',
        summary: 'Ошибка',
        detail: err.data?.error || err.error || 'Не удалось создать папку',
        life: 5000,
      })
    }
  }

  const renderBreadcrumb = () => {
    if (!currentPath) return null
    const parts = currentPath.split('/').filter(Boolean)
    return (
      <div className="flex align-items-center gap-1 mb-3 flex-wrap">
        <Button
          label="Корень"
          icon="pi pi-home"
          onClick={() => onNavigate('')}
          text
          size="small"
        />
        <span className="text-gray-400">/</span>
        {parts.map((part, index) => {
          const pathSoFar = parts.slice(0, index + 1).join('/')
          return (
            <React.Fragment key={pathSoFar}>
              <Button
                label={part}
                onClick={() => onNavigate(pathSoFar)}
                text
                size="small"
              />
              {index < parts.length - 1 && <span className="text-gray-400">/</span>}
            </React.Fragment>
          )
        })}
      </div>
    )
  }

  return (
    <div>
      <Toast ref={toast} />

      {/* Breadcrumb */}
      {renderBreadcrumb()}

      {/* Files Table */}
      <div className="p-card">
        <DataTable
          value={files}
          loading={isLoading}
          responsiveLayout="scroll"
          selection={selectedFile}
          onSelectionChange={(e: any) => setSelectedFile(e.value)}
          rowClassName={(file: FileEntry) =>
            file.type === 'folder' ? 'cursor-pointer hover:bg-blue-50' : 'cursor-pointer hover:bg-gray-50'
          }
          onRowDoubleClick={(e: DataTableRowClickEvent) => handleNavigate(e.data as FileEntry)}
          dataKey="name"
        >
          <Column
            header=""
            body={(file: FileEntry) => (
              <i className={getFileIcon(file)}></i>
            )}
            style={{ width: '60px', textAlign: 'center' }}
          />
          <Column
            field="name"
            header="Имя"
            body={(file: FileEntry) => (
              <span className="font-medium">{file.name}</span>
            )}
            style={{ width: '40%' }}
          />
          <Column
            field="type"
            header="Тип"
            body={(file: FileEntry) => (
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                file.type === 'folder'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-blue-100 text-blue-800'
              }`}>
                {file.type === 'folder' ? 'Папка' : 'Файл'}
              </span>
            )}
            style={{ width: '15%' }}
          />
          <Column
            field="size"
            header="Размер"
            body={(file: FileEntry) => formatSize(file.size)}
            style={{ width: '15%' }}
          />
          <Column
            field="modified"
            header="Изменён"
            body={(file: FileEntry) => formatDate(file.modified)}
            style={{ width: '20%' }}
          />
        </DataTable>
      </div>

      {/* Upload Dialog */}
      <Dialog
        visible={uploadVisible}
        onHide={() => !isUploading && setUploadVisible(false)}
        header="Загрузка файла"
        style={{ width: '500px' }}
      >
        <FileUpload
          mode="basic"
          chooseLabel="Выбрать файл"
          uploadLabel="Загрузить"
          cancelLabel="Отмена"
          auto
          chooseOptions={{ icon: 'pi pi-search' }}
          uploadHandler={handleUpload}
          maxFileSize={100000000}
          emptyTemplate={
            <p className="mt-4">
              Перетащите файл сюда для загрузки или нажмите для выбора файла
            </p>
          }
        />
      </Dialog>

      {/* Create Directory Dialog */}
      <Dialog
        visible={mkdirVisible}
        onHide={() => !isCreating && setMkdirVisible(false)}
        header="Создать папку"
        style={{ width: '400px' }}
        footer={
          <div>
            <Button
              label="Отмена"
              icon="pi pi-times"
              onClick={() => setMkdirVisible(false)}
              text
              disabled={isCreating}
            />
            <Button
              label="Создать"
              icon="pi pi-check"
              onClick={handleCreateDirectory}
              disabled={!newDirName.trim() || isCreating}
            />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <label>Название папки</label>
          <InputText
            value={newDirName}
            onChange={(e) => setNewDirName(e.target.value)}
            placeholder="Введите название папки"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreateDirectory()
            }}
          />
        </div>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        visible={!!selectedFile}
        onHide={() => setSelectedFile(null)}
        header="Подтверждение удаления"
        style={{ width: '400px' }}
        footer={
          <div>
            <Button
              label="Отмена"
              icon="pi pi-times"
              onClick={() => setSelectedFile(null)}
              text
              disabled={isDeleting}
            />
            <Button
              label="Удалить"
              icon="pi pi-trash"
              onClick={handleDelete}
              severity="danger"
              disabled={isDeleting}
            />
          </div>
        }
      >
        <p>
          Вы действительно хотите удалить <strong>{selectedFile?.name}</strong>?
        </p>
      </Dialog>
    </div>
  )
}
