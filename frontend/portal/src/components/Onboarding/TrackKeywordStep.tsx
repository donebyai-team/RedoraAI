"use client";

import {
    TextField,
    Typography,
    Chip,
    Box,
    InputAdornment,
    IconButton,
    Paper,
    Stack,
} from "@mui/material";
import { Plus } from "lucide-react";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import {
    nextStep,
    prevStep,
    resetStepper,
    setProjects,
} from "../../../store/Onboarding/OnboardingSlice";
import { useState, useCallback } from "react";
import { useForm, Controller } from "react-hook-form";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";

interface TrackKeywordFormValues {
    keywords: string[];
    newKeyword: string;
}

export default function TrackKeywordStep() {
    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const projects = useAppSelector((state: RootState) => state.stepper.projects);
    const listOfSuggestedKeywords = projects?.suggestedKeywords ?? [];
    const { portalClient } = useClientsContext();

    const [isLoading, setIsLoading] = useState(false);

    const {
        handleSubmit,
        control,
        watch,
        setValue,
    } = useForm<TrackKeywordFormValues>({
        defaultValues: {
            keywords: projects?.keywords ?? [],
            newKeyword: "",
        },
    });

    const keywords = watch("keywords");
    const newKeyword = watch("newKeyword");

    const handleAddKeyword = useCallback(() => {
        const trimmed = newKeyword.trim();
    
        if (!trimmed) return;
    
        const isDuplicate = keywords.some(
            (k) => k.toLowerCase() === trimmed.toLowerCase()
        );
    
        if (isDuplicate) {
            toast.error(`"${trimmed}" is already added`);
            return;
        }
    
        setValue("keywords", [...keywords, trimmed]);
        setValue("newKeyword", "");
    }, [newKeyword, keywords, setValue]);
    

    const handleDeleteKeyword = useCallback((index: number) => {
        const updated = keywords.filter((_, i) => i !== index);
        setValue("keywords", updated);
    }, [keywords, setValue]);

    const onSubmit = useCallback(
        async (data: TrackKeywordFormValues) => {
            if (!projects) return;
            setIsLoading(true);

            try {
                await portalClient.createKeywords({ keywords: data.keywords });

                dispatch(setProjects({ ...projects, keywords: data.keywords }));

                toast.success("Keywords saved successfully");
                dispatch(nextStep());
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong";
                toast.error(message);
            } finally {
                setIsLoading(false);
            }
        },
        [dispatch, portalClient, projects]
    );

    const handleBack = useCallback(() => dispatch(prevStep()), [dispatch]);
    const handleReset = useCallback(() => dispatch(resetStepper()), [dispatch]);

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3} mb={5}>
                <Controller
                    name="newKeyword"
                    control={control}
                    render={({ field }) => (
                        <TextField
                            fullWidth
                            label="Add Keyword"
                            {...field}
                            onKeyDown={(e) => {
                                if (e.key === "Enter") {
                                    e.preventDefault();
                                    handleAddKeyword();
                                }
                            }}
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
                    )}
                />

                <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                    {keywords.map((keyword, index) => (
                        <Chip
                            key={`${keyword}-${index}`}
                            label={keyword}
                            onDelete={() => handleDeleteKeyword(index)}
                            color="primary"
                            variant="outlined"
                        />
                    ))}
                </Box>

                {listOfSuggestedKeywords.length > 0 && (
                    <Paper variant="outlined" sx={{ p: 2, bgcolor: "background.default" }}>
                        <Typography variant="subtitle2" gutterBottom>
                            Suggested Keywords
                        </Typography>
                        <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                            {listOfSuggestedKeywords.map((suggestion, index) => (
                                <Chip
                                    key={index}
                                    label={suggestion}
                                    onClick={() => setValue("newKeyword", suggestion)}
                                    size="small"
                                />
                            ))}
                        </Stack>
                    </Paper>
                )}
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
    );
}
