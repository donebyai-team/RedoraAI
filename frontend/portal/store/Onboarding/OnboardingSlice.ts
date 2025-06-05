import { Project } from '@doota/pb/doota/core/v1/core_pb'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

interface StepperState {
  currentStep: number;
  completedSteps: number[];
  project: Project | null;
  isOnboardingDone: boolean | null;
  loading: boolean;
}

const initialState: StepperState = {
  currentStep: 1,
  completedSteps: [],
  project: null,
  isOnboardingDone: null,
  loading: false
}

const stepperSlice = createSlice({
  name: 'stepper',
  initialState,
  reducers: {
    goToStep(state, action: PayloadAction<number>) {
      const targetStep = action.payload;

      for (let step = 1; step < targetStep; step++) {
        if (!state.completedSteps.includes(step)) {
          state.completedSteps.push(step);
        }
      }

      state.currentStep = targetStep;
    },
    nextStep(state) {
      if (!state.completedSteps.includes(state.currentStep)) {
        state.completedSteps.push(state.currentStep);
      }
      state.currentStep += 1;
    },
    prevStep(state) {
      if (state.currentStep > 1) state.currentStep -= 1;
    },
    finishOnboarding(state) {
      if (!state.completedSteps.includes(state.currentStep)) {
        state.completedSteps.push(state.currentStep);
      }
    },
    setProject: (state, action: PayloadAction<Project>) => {
      state.project = action.payload
    },
    setIsOnboardingDone: (state, action: PayloadAction<boolean | null>) => {
      state.isOnboardingDone = action.payload
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    }
  }
})

export const {
  goToStep,
  nextStep,
  prevStep,
  finishOnboarding,
  setProject,
  setIsOnboardingDone,
  setLoading,
} = stepperSlice.actions

export const stepperReducer = stepperSlice.reducer
