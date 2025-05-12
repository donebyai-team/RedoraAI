import React from 'react';
import {
    Box,
    Typography,
} from '@mui/material';
import ProductInformationStep from './ProductInformationStep';
import TrackKeywordStep from './TrackKeywordStep';
import SelectSourcesStep from './SelectSourcesStep';

interface StepContentProps {
    step: number;
    stepLength?: number
}

const StepContent: React.FC<StepContentProps> = ({ step, stepLength }) => {

    const StepCounter = () => (
        <Typography
            variant="body2"
            sx={{
                color: 'primary.main',
                fontWeight: 500,
                mb: 2,
                display: 'block'
            }}
        >
            Step {step + 1} of {stepLength}
        </Typography>
    );

    switch (step) {
        case 0:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Product Information
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Tell us about your product to help us track relevant discussions.
                    </Typography>

                    <ProductInformationStep />
                </Box>
            );

        case 1:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Track Keywords
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Add keywords related to your product that you want to track on Reddit.
                    </Typography>

                    <TrackKeywordStep />
                </Box>
            );

        case 2:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Select Sources
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Choose your source where you want to track your keywords.
                    </Typography>

                    <SelectSourcesStep />
                </Box>
            );

        default:
            return null;
    }
};

export default StepContent;