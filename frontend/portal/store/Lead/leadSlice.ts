import { Lead } from "@doota/pb/doota/core/v1/core_pb";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export type LeadTyeps = Lead;

// Define the types
interface LeadStateTyeps {
    listofleads: LeadTyeps[];
    selectedleadData: LeadTyeps | null;
    isLoading: boolean;
    error: string | null;
}

// Initial state
const initialState: LeadStateTyeps = {
    listofleads: [],
    selectedleadData: null,
    isLoading: false,
    error: null,
};

// Slice
const leadSlice = createSlice({
    name: 'lead',
    initialState,
    reducers: {
        setListOfLeads: (state, action: PayloadAction<LeadTyeps[]>) => {
            state.listofleads = action.payload;
        },
        setSelectedLeadData: (state, action: PayloadAction<LeadTyeps | null>) => {
            state.selectedleadData = action.payload;
        },
        setIsLoading: (state, action: PayloadAction<boolean>) => {
            state.isLoading = action.payload;
        },
        setError: (state, action: PayloadAction<string>) => {
            state.error = action.payload;
        },
    },

});

// Export actions
export const { setListOfLeads, setSelectedLeadData, setError, setIsLoading } = leadSlice.actions;

// Export reducer
export const leadReducer = leadSlice.reducer;
