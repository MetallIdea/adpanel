import { configureStore } from '@reduxjs/toolkit'
import { sitesReducer } from './sitesSlice'
import { sitesApi } from '../services/sitesApi'

export const store = configureStore({
  reducer: {
    sites: sitesReducer,
    [sitesApi.reducerPath]: sitesApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(sitesApi.middleware),
})

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch
