"use client";

import {
    TextField,
    Typography,
    Chip
} from "@mui/material";
import { Stack } from "@mui/system";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import { nextStep, prevStep, resetStepper, setProjects } from "../../../store/Onboarding/OnboardingSlice";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import {
    Box,
    InputAdornment,
    IconButton,
    Paper,
} from '@mui/material';
import { Plus } from 'lucide-react';
import { Keyword } from "@doota/pb/doota/core/v1/core_pb";

interface TrackKeywordFormValues {
    keywords: Keyword[];
}

export default function TrackKeywordStep() {

    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const projects = useAppSelector((state: RootState) => state.stepper.projects);
    const { portalClient } = useClientsContext();

    const {
        handleSubmit,
    } = useForm<TrackKeywordFormValues>({
        defaultValues: {
            keywords: projects?.keywords ?? [],
        },
    });

    const [isLoading, setIsLoading] = useState(false)

    const onSubmit = async (data: TrackKeywordFormValues) => {
        setIsLoading(true);

        try {
            const body = {
                ...(projects?.id && { id: projects.id }),
                ...data,
            };

            const result = await portalClient.createOrEditProject(body);
            dispatch(setProjects(result));

            toast.success("Keywords saved successfully");
            dispatch(nextStep());
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    };

    const handleBack = () => {
        dispatch(prevStep());
    };

    const handleReset = () => {
        dispatch(resetStepper());
    };

    const [keywords, setKeywords] = useState<string[]>([]);
    const [newKeyword, setNewKeyword] = useState('');

    const handleAddKeyword = () => {
        if (newKeyword && !keywords.includes(newKeyword)) {
            setKeywords([...keywords, newKeyword]);
            setNewKeyword('');
        }
    };

    const handleDeleteKeyword = (keywordToDelete: string) => {
        setKeywords(keywords.filter(keyword => keyword !== keywordToDelete));
    };

    return (<>
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3}>
                <TextField
                    fullWidth
                    label="Add Keyword"
                    value={newKeyword}
                    onChange={(e) => setNewKeyword(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleAddKeyword()}
                    InputProps={{
                        endAdornment: (
                            <InputAdornment position="end">
                                <IconButton onClick={handleAddKeyword} edge="end">
                                    <Plus size={20} />
                                </IconButton>
                            </InputAdornment>
                        ),
                    }}
                />

                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                    {keywords.map((keyword) => (
                        <Chip
                            key={keyword}
                            label={keyword}
                            onDelete={() => handleDeleteKeyword(keyword)}
                            color="primary"
                            variant="outlined"
                        />
                    ))}
                </Box>

                <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                    <Typography variant="subtitle2" gutterBottom>
                        Suggested Keywords
                    </Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                        {['competitor', 'alternative', 'vs', 'review', 'help'].map((suggestion, index) => (
                            <Chip
                                key={index}
                                label={suggestion}
                                onClick={() => setNewKeyword(suggestion)}
                                size="small"
                            />
                        ))}
                    </Stack>
                </Paper>
            </Stack>

            <StepperControls
                activeStep={activeStep}
                handleBack={handleBack}
                handleNext={handleSubmit(onSubmit)}
                handleReset={handleReset}
                steps={steps}
                btnDisabled={isLoading}
            />
        </form>
    </>);
}