import { Project } from '@doota/pb/doota/core/v1/core_pb'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export type ProjectsTypes = Project | null

export type SourcesTypes = {
  id: string
  name: string
}

type ProjectTypes = {
  id?: string
  name: string
  description: string
  website: string
  targetPersona: string
  keywords?: string[]
  sources?: SourcesTypes[]
  suggestedKeywords?: string[]
  suggestedSources?: string[]
}

interface StepperState {
  activeStep: number
  skipped: number[]
  projects: ProjectTypes | null
}

const initialState: StepperState = {
  activeStep: 0,
  skipped: [],
  projects: null
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
    setProjects: (state, action: PayloadAction<ProjectTypes>) => {
      state.projects = action.payload
    }
  }
})

export const { nextStep, prevStep, resetStepper, skipStep, setStep, setProjects } = stepperSlice.actions

export const stepperReducer = stepperSlice.reducer
