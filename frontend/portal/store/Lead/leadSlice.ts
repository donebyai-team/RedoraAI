import { Lead, LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { DateRangeFilter } from "@doota/pb/doota/portal/v1/portal_pb";

export type LeadTyeps = Lead;

export enum LeadTabStatus {
    NEW = "new",
    COMPLETED = "completed",
    DISCARDED = "discarded",
    LEAD = "lead"
}

// Define the types
interface LeadStateTyeps {
    newTabList: LeadTyeps[];
    completedTabList: LeadTyeps[];
    discardedTabList: LeadTyeps[];
    leadsTabList: LeadTyeps[];
    selectedleadData: LeadTyeps | null;
    isLoading: boolean;
    activeTab: LeadTabStatus;
    error: string | null;
    leadStatusFilter: LeadStatus | null;
    dateRange: DateRangeFilter;
}

// Initial state
const initialState: LeadStateTyeps = {
    newTabList: [],
    completedTabList: [],
    discardedTabList: [],
    leadsTabList: [],
    selectedleadData: null,
    isLoading: true,
    activeTab: LeadTabStatus.NEW,
    error: null,
    leadStatusFilter: null,
    dateRange: DateRangeFilter.DATE_RANGE_7_DAYS
};

// Slice
const leadSlice = createSlice({
    name: 'lead',
    initialState,
    reducers: {
        setNewTabList: (state, action: PayloadAction<LeadTyeps[]>) => {
            state.newTabList = action.payload;
        },
        setCompletedList: (state, action: PayloadAction<LeadTyeps[]>) => {
            state.completedTabList = action.payload;
        },
        setDiscardedTabList: (state, action: PayloadAction<LeadTyeps[]>) => {
            state.discardedTabList = action.payload;
        },
        setLeadsTabList: (state, action: PayloadAction<LeadTyeps[]>) => {
            state.leadsTabList = action.payload;
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
        setLeadStatusFilter: (state, action: PayloadAction<LeadStatus | null>) => {
            state.leadStatusFilter = action.payload;
        },
        setDateRange: (state, action: PayloadAction<DateRangeFilter>) => {
            state.dateRange = action.payload;
        },
    },

});

// Export actions
export const {
    setNewTabList,
    setCompletedList,
    setDiscardedTabList,
    setLeadsTabList,
    setSelectedLeadData,
    setActiveTab,
    setError,
    setIsLoading,
    setLeadStatusFilter,
    setDateRange,
} = leadSlice.actions;

// Export reducer
export const leadReducer = leadSlice.reducer;
