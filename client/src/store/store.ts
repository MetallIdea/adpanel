import { configureStore } from '@reduxjs/toolkit'
import { sitesReducer } from './sitesSlice'
import { sitesApi } from '../services/sitesApi'
import { servicesReducer } from './servicesSlice'
import { servicesApi } from '../services/servicesApi'
import { filesReducer } from './filesSlice'
import { filesApi } from '../services/filesApi'

export const store = configureStore({
  reducer: {
    sites: sitesReducer,
    [sitesApi.reducerPath]: sitesApi.reducer,
    services: servicesReducer,
    [servicesApi.reducerPath]: servicesApi.reducer,
    files: filesReducer,
    [filesApi.reducerPath]: filesApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(sitesApi.middleware)
      .concat(servicesApi.middleware)
      .concat(filesApi.middleware),
})

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch
