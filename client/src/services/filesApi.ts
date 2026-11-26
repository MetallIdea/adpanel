import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import { API_BASE_URL } from '../constants/api'

export interface FileEntry {
  name: string
  type: 'file' | 'folder'
  size: number
  path: string
  modified: string
}

export interface BreadcrumbItem {
  label: string
  path: string
}

export const filesApi = createApi({
  reducerPath: 'filesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getFiles: builder.query<FileEntry[], { path?: string }>({
      query: ({ path } = {}) => ({
        url: '/files',
        params: path ? { path } : undefined,
      }),
    }),
    uploadFile: builder.mutation<{ message: string; name: string; path: string; size: number }, { file: File; path?: string }>({
      query: ({ file, path }) => {
        const formData = new FormData()
        formData.append('file', file)
        if (path) formData.append('path', path)
        return {
          url: '/files/upload',
          method: 'POST',
          body: formData,
        }
      },
    }),
    deleteFile: builder.mutation<{ message: string }, string>({
      query: (path) => ({
        url: `/files/${path}`,
        method: 'DELETE',
      }),
    }),
    createDirectory: builder.mutation<{ message: string; path: string }, { path: string; name: string }>({
      query: ({ path, name }) => ({
        url: '/files/mkdir',
        method: 'POST',
        body: { path, name },
      }),
    }),
    getBreadcrumb: builder.query<BreadcrumbItem[], { path?: string }>({
      query: ({ path } = {}) => ({
        url: '/files/breadcrumb',
        params: path ? { path } : undefined,
      }),
    }),
  }),
  tagTypes: ['File'],
})

export const {
  useGetFilesQuery,
  useUploadFileMutation,
  useDeleteFileMutation,
  useCreateDirectoryMutation,
  useGetBreadcrumbQuery,
} = filesApi
