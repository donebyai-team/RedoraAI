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
import { useAppSelector } from '../../../store/hooks';
import { RootState } from '../../../store/store';
import { AuthLoading } from '../../app/(restricted)/dashboard/layout';
import { isActivePath } from '../../utils/url';
import { routes } from '@doota/ui-core/routing';
import { usePathname } from 'next/navigation';

export const steps = [
    {
        label: 'Product Details',
        description: 'Your business details in brief'
    },
    {
        label: 'Track Keywords',
        description: 'Choose keywords to track.'
    },
    {
        label: 'Select SubReddits',
        description: 'Subreddits you want to monitor.'
    },
];

export default function OnboadingForm() {

    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const pathname = usePathname();
    const { skipped, loading } = useAppSelector((state: RootState) => state.stepper);
    const isStepSkipped = (step: number) => skipped.includes(step);
    const isEditProduct = isActivePath(routes.app.settings.edit_product, pathname);

    if (loading) {
        return <AuthLoading />
    }

    return (<>
        <Box sx={{ width: "100%", height: isEditProduct ? "90vh" : "100%", ...(isEditProduct && { display: "flex", flexDirection: "column", justifyContent: "center" }) }}>
            <Box
                className={isEditProduct ? "" : "min-h-screen flex flex-col justify-center"}
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
                                            minHeight: '60px',
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
