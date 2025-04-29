import { Lead } from "@doota/pb/doota/core/v1/core_pb";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export type LeadTyeps = Lead;

export enum LeadTabStatus {
    NEW = "new",
    COMPLETED = "completed",
    DISCARDED = "discarded",
}

// Define the types
interface LeadStateTyeps {
    listofleads: LeadTyeps[];
    selectedleadData: LeadTyeps | null;
    isLoading: boolean;
    activeTab: LeadTabStatus;
    error: string | null;
}

// Initial state
const initialState: LeadStateTyeps = {
    listofleads: [],
    selectedleadData: null,
    isLoading: false,
    activeTab: LeadTabStatus.NEW,
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
        setActiveTab: (state, action: PayloadAction<LeadTabStatus>) => {
            state.activeTab = action.payload;
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
export const { setListOfLeads, setSelectedLeadData, setActiveTab, setError, setIsLoading } = leadSlice.actions;

// Export reducer
export const leadReducer = leadSlice.reducer;
