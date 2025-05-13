'use client';

import { configureStore, combineReducers } from '@reduxjs/toolkit';
import { persistReducer, persistStore } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import { sourceReducer } from './Source/sourceSlice';
import { leadReducer } from './Lead/leadSlice';
import { paremsReducer } from './Params/ParamsSlice';
import { redditIntegrationReducer } from './slices/redditIntegrationSlice';
import { stepperReducer } from './Onboarding/OnboardingSlice';

const rootReducer = combineReducers({
  source: sourceReducer,
  lead: leadReducer,
  parems: paremsReducer,
  redditIntegration: redditIntegrationReducer,
  stepper: stepperReducer,
});

const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['parems']
};

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer: persistedReducer,
  middleware: getDefaultMiddleware =>
    getDefaultMiddleware({
      serializableCheck: false,
    }),
});

export const persistor = persistStore(store);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;