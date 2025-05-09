import { Source } from "@doota/pb/doota/core/v1/core_pb";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export type SourceTyeps = Source;

interface ModifySourceTyeps extends SourceTyeps {
    leads_count?: number;
}

// Define the types
interface SourceStateTyeps {
    subredditList: ModifySourceTyeps[];
    loading: boolean;
    error: string | null;
}

// Initial state
const initialState: SourceStateTyeps = {
    subredditList: [],
    loading: false,
    error: null,
};

// Slice
const sourceSlice = createSlice({
    name: 'source',
    initialState,
    reducers: {
        setSubredditList: (state, action: PayloadAction<SourceTyeps[]>) => {
            state.subredditList = action.payload;
        },
        setLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload;
        },
        setError: (state, action: PayloadAction<string>) => {
            state.error = action.payload;
        },
    },

});

// Export actions
export const { setSubredditList, setError, setLoading } = sourceSlice.actions;

// Export reducer
export const sourceReducer = sourceSlice.reducer;
