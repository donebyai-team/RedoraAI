"use client"

import {
    Box,
    Stepper,
    Step,
    StepLabel,
    Paper,
    Grid,
    Typography,
} from '@mui/material';
import { Container } from '@mui/system';
import StepContent from './StepContent';
import CustomStepIcon from './CustomStepIcon';
import { useOnboardingStatus } from '../../hooks/useOnboardingStatus';
import { useAppSelector } from '../../../store/hooks';
import { RootState } from '../../../store/store';
import { AuthLoading } from '../../app/(restricted)/dashboard/layout';
import { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { setProjects, skipStep } from '../../../store/Onboarding/OnboardingSlice';

export const steps = [
    {
        label: 'Product Information',
        description: 'Enter your product details and basic information.'
    },
    {
        label: 'Track Keywords',
        description: 'Choose keywords to track across Reddit.'
    },
    {
        label: 'Select Sources',
        description: 'Select sources to monitor for your keywords.'
    },
    // {
    //     label: 'Connect Reddit',
    //     description: 'Connect your Reddit account to start tracking.'
    // }
];

export default function ManinForm() {

    const { loading, data, error } = useOnboardingStatus();
    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const skipped = useAppSelector((state: RootState) => state.stepper.skipped);
    const [isLoading, setIsLoading] = useState<boolean>(false);

    console.log("###_debug_data ", { loading, data });

    const isStepSkipped = (step: number) => skipped.includes(step);

    useEffect(() => {
        const handleStepNavigation = async () => {
            if (loading || error || !data) return;

            setIsLoading(true);

            // Step 1: Set projects in store
            dispatch(setProjects(data));

            let nextStep = 0;

            // Step 2: Check if project data is valid
            if (data !== null) {
                const { id, website, name, description, targetPersona, keywords, sources } = data;

                // Wait until basic info is verified
                const hasBasicInfo = Boolean(id && website && name && description && targetPersona);
                await new Promise((resolve) => setTimeout(resolve, 0)); // micro-wait for async flow

                if (hasBasicInfo) {
                    nextStep = 1;

                    // Step 3: Check keywords
                    const hasKeywords = Array.isArray(keywords) && keywords.length > 0;
                    await new Promise((resolve) => setTimeout(resolve, 0)); // micro-wait again

                    if (hasKeywords) {
                        nextStep = 2;

                        // Step 4: Check sources
                        const hasSources = Array.isArray(sources) && sources.length > 0;
                        await new Promise((resolve) => setTimeout(resolve, 0));

                        if (hasSources) {
                            nextStep = 3;
                        }
                    }
                }
            }

            // Step 5: Finally, set active step in store
            dispatch(skipStep(nextStep));

            setIsLoading(false);
        };

        handleStepNavigation();
    }, [loading, data, error, dispatch]);

    if (loading || isLoading) {
        return <AuthLoading />
    }

    return (<>
        <Box sx={{ width: "100%", height: "100%" }}>
            <Box
                className="min-h-screen flex flex-col justify-center"
                sx={{ py: { xs: 2, sm: 4, md: 6 }, background: '#f9fafb' }}
            >
                <Container maxWidth="md">
                    <Paper
                        elevation={3}
                        sx={{
                            minHeight: '400px',
                            display: 'flex',
                            flexDirection: 'column',
                            padding: 10,
                            borderRadius: 7,
                            boxShadow: "0 4px 20px rgba(0, 0, 0, 0.08)",
                            overflow: "hidden",
                            position: "relative",
                        }}
                    >
                        <Grid container spacing={4}>
                            <Grid item xs={12} md={4}>
                                <Stepper
                                    activeStep={activeStep}
                                    orientation="vertical"
                                    sx={{
                                        '& .MuiStepConnector-line': {
                                            minHeight: '40px',
                                            borderLeftWidth: '2px',
                                            ml: '8px',
                                            borderColor: 'rgba(25, 118, 210, 0.2)'
                                        },
                                        '& .MuiStepLabel-root': {
                                            padding: '8px 0'
                                        },
                                        '& .MuiStepIcon-root': {
                                            fontSize: '2rem'
                                        }
                                    }}
                                >
                                    {steps.map((step, index) => {
                                        const stepProps: { completed?: boolean } = {};
                                        if (isStepSkipped(index)) {
                                            stepProps.completed = false;
                                        }

                                        return (
                                            <Step key={step.label} {...stepProps}>
                                                <StepLabel
                                                    StepIconComponent={CustomStepIcon}
                                                    sx={{
                                                        '& .MuiStepLabel-label': {
                                                            color: activeStep === index ? 'primary.main' : 'text.primary',
                                                            fontWeight: activeStep === index ? 600 : 400
                                                        }
                                                    }}
                                                >
                                                    <Typography
                                                        variant="subtitle1"
                                                        sx={{
                                                            mb: 0.5,
                                                            lineHeight: 1.2
                                                        }}
                                                    >
                                                        {step.label}
                                                    </Typography>
                                                    <Typography
                                                        variant="body2"
                                                        color="text.secondary"
                                                        sx={{
                                                            display: 'block',
                                                            maxWidth: '240px',
                                                            lineHeight: 1.4
                                                        }}
                                                    >
                                                        {step.description}
                                                    </Typography>
                                                </StepLabel>
                                            </Step>
                                        );
                                    })}
                                </Stepper>
                            </Grid>

                            <Grid item xs={12} md={8}>
                                <Box sx={{
                                    height: '100%',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    pl: { md: 4 },
                                    borderLeft: { md: '1px solid rgba(0, 0, 0, 0.08)' }
                                }}>
                                    <Box sx={{ flex: 1 }}>
                                        <StepContent step={activeStep} stepLength={steps.length} />
                                    </Box>


                                </Box>
                            </Grid>
                        </Grid>
                    </Paper>
                </Container>
            </Box>
        </Box>
    </>);
}
