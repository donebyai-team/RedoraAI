import { createSlice, PayloadAction } from "@reduxjs/toolkit";

// Define the types
interface ParamsStateTyeps {
    relevancyScore: number;
    subReddit: string;
}

// Initial state
const initialState: ParamsStateTyeps = {
    relevancyScore: 70,
    subReddit: "",
};

// Slice
const paramsSlice = createSlice({
    name: 'params',
    initialState,
    reducers: {
        setRelevancyScore: (state, action: PayloadAction<number>) => {
            state.relevancyScore = action.payload;
        },
        setSubReddit: (state, action: PayloadAction<string>) => {
            state.subReddit = action.payload;
        },
    },

});

// Export actions
export const { setRelevancyScore, setSubReddit } = paramsSlice.actions;

// Export reducer
export const paremsReducer = paramsSlice.reducer;
