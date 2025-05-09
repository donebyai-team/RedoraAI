import { createSlice, PayloadAction } from '@reduxjs/toolkit';

type IntegrationState = {
    isConnected: boolean | null;
    loading: boolean;
    error: string | null;
};

const initialState: IntegrationState = {
    isConnected: null,
    loading: false,
    error: null,
};

export const redditIntegrationSlice = createSlice({
    name: 'redditIntegration',
    initialState,
    reducers: {
        setLoading(state) {
            state.loading = true;
            state.error = null;
        },
        setSuccess(state, action: PayloadAction<boolean>) {
            state.isConnected = action.payload;
            state.loading = false;
            state.error = null;
        },
        setError(state, action: PayloadAction<string>) {
            state.isConnected = false;
            state.loading = false;
            state.error = action.payload;
        },
    },
});

export const { setLoading, setSuccess, setError } = redditIntegrationSlice.actions;

// Export reducer
export const redditIntegrationReducer = redditIntegrationSlice.reducer;
