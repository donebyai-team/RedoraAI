import { Lead, LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { DateRangeFilter, LeadAnalysis } from "@doota/pb/doota/portal/v1/portal_pb";

// Define the types
interface LeadStateTyeps {
    leadList: Lead[];
    selectedleadData: Lead | null;
    isLoading: boolean;
    error: string | null;
    leadStatusFilter: LeadStatus | null;
    dateRange: DateRangeFilter;
    dashboardCounts: LeadAnalysis | undefined;
}

// Initial state
const initialState: LeadStateTyeps = {
    leadList: [],
    selectedleadData: null,
    isLoading: false,
    error: null,
    leadStatusFilter: null,
    dateRange: DateRangeFilter.DATE_RANGE_7_DAYS,
    dashboardCounts: undefined
};

// Slice
const leadSlice = createSlice({
    name: 'lead',
    initialState,
    reducers: {
        setLeadList: (state, action: PayloadAction<Lead[]>) => {
            state.leadList = action.payload;
        },
        setDashboardCounts: (state, action: PayloadAction<LeadAnalysis | undefined>) => {
            state.dashboardCounts = action.payload;
        },
        setSelectedLeadData: (state, action: PayloadAction<Lead | null>) => {
            state.selectedleadData = action.payload;
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
    setLeadList,
    setDashboardCounts,
    setSelectedLeadData,
    setError,
    setIsLoading,
    setLeadStatusFilter,
    setDateRange,
} = leadSlice.actions;

// Export reducer
export const leadReducer = leadSlice.reducer;
