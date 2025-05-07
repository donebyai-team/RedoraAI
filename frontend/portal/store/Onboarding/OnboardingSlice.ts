import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface StepperState {
    activeStep: number;
    skipped: number[];
}

const initialState: StepperState = {
    activeStep: 0,
    skipped: [],
};

const stepperSlice = createSlice({
    name: 'stepper',
    initialState,
    reducers: {
        // Replace Set operations with array logic:
        nextStep: (state) => {
            state.skipped = state.skipped.filter(step => step !== state.activeStep);
            state.activeStep += 1;
        },
        skipStep: (state, action: PayloadAction<number>) => {
            if (!state.skipped.includes(action.payload)) {
                state.skipped.push(action.payload);
            }
        },
        prevStep: (state) => {
            state.activeStep = Math.max(0, state.activeStep - 1);
        },
        resetStepper: (state) => {
            state.activeStep = 0;
            state.skipped = [];
        },
        setStep: (state, action: PayloadAction<number>) => {
            state.activeStep = action.payload;
        },
    },
});

export const {
    nextStep,
    prevStep,
    resetStepper,
    skipStep,
    setStep,
} = stepperSlice.actions;

export const stepperReducer = stepperSlice.reducer;
