import { configureStore } from '@reduxjs/toolkit';
import { sourceReducer } from './Source/sourceSlice';
import { leadReducer } from './Lead/leadSlice';
import { paremsReducer } from './Params/ParamsSlice';

export const store = configureStore({
  reducer: {
    source: sourceReducer,
    lead: leadReducer,
    parems: paremsReducer
  },
});

// 👇 export RootState type
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;