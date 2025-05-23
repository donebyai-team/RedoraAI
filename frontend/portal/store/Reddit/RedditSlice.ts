import { Integration } from '@doota/pb/doota/portal/v1/portal_pb';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface RedditState {
    accounts: Integration[];
    loading: boolean;
    error: string | null;
}

const initialState: RedditState = {
    accounts: [],
    loading: false,
    error: null,
};

const redditSlice = createSlice({
    name: 'reddit',
    initialState,
    reducers: {
        setAccounts(state, action: PayloadAction<Integration[]>) {
            state.accounts = action.payload;
        },
        setLoading(state, action: PayloadAction<boolean>) {
            state.loading = action.payload;
        },
        setError(state, action: PayloadAction<string | null>) {
            state.error = action.payload;
        },
    }
});

export const { setAccounts, setLoading, setError } = redditSlice.actions;

export const redditReducer = redditSlice.reducer;
