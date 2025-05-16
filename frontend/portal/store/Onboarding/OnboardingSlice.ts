import { Project } from '@doota/pb/doota/core/v1/core_pb'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

interface StepperState {
  activeStep: number
  skipped: number[]
  project: Project | null
  isOnboardingDone: boolean | null
  loading: boolean
}

const initialState: StepperState = {
  activeStep: 0,
  skipped: [],
  project: null,
  isOnboardingDone: null,
  loading: false
}

const stepperSlice = createSlice({
  name: 'stepper',
  initialState,
  reducers: {
    // Replace Set operations with array logic:
    nextStep: state => {
      state.skipped = state.skipped.filter(step => step !== state.activeStep)
      state.activeStep += 1
    },
    skipStep: (state, action: PayloadAction<number>) => {
      if (!state.skipped.includes(action.payload)) {
        state.skipped.push(action.payload)
      }
    },
    prevStep: state => {
      state.activeStep = Math.max(0, state.activeStep - 1)
    },
    resetStepper: state => {
      state.activeStep = 0
      state.skipped = []
    },
    setStep: (state, action: PayloadAction<number>) => {
      state.activeStep = action.payload
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

export const { nextStep, prevStep, resetStepper, skipStep, setStep, setProject, setIsOnboardingDone, setLoading } =
  stepperSlice.actions

export const stepperReducer = stepperSlice.reducer
