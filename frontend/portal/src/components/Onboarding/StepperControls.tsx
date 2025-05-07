import React from 'react';
import { Box, Button } from '@mui/material';
import { ArrowLeft, ArrowRight, Check } from 'lucide-react';

interface StepperControlsProps {
  activeStep: number;
  handleBack: () => void;
  handleNext: () => void;
  handleReset: () => void;
  steps: { label: string; description: string; }[];
  btnDisabled: boolean;
}

const StepperControls: React.FC<StepperControlsProps> = ({
  activeStep,
  handleBack,
  handleNext,
  handleReset,
  steps,
  btnDisabled
}) => {
  //   const isLastStep = activeStep === steps.length - 1;
  const isFirstStep = btnDisabled || activeStep === 0;

  return (
    <Box sx={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', pt: 2, pb: 2 }}>
      <Button
        color="inherit"
        variant="outlined"
        disabled={isFirstStep}
        onClick={handleBack}
        startIcon={<ArrowLeft size={18} />}
        sx={{ mr: 1 }}
      >
        Back
      </Button>
      <Box sx={{ flex: '1 1 auto' }} />
      {activeStep === steps.length - 1 ? (
        <Button
          variant="contained"
          color="success"
          onClick={handleReset}
          endIcon={<Check size={18} />}
          disabled={btnDisabled}
        >
          Finish
        </Button>
      ) : (
        <Button
          variant="contained"
          onClick={handleNext}
          endIcon={<ArrowRight size={18} />}
          disabled={btnDisabled}
        >
          Next
        </Button>
      )}
    </Box>
  );
};

export default StepperControls;