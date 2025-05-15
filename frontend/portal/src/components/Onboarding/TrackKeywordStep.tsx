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
    setProject,
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
    const project = useAppSelector((state: RootState) => state.stepper.project);
    const listOfSuggestedKeywords = project?.suggestedKeywords ?? [];
    const { portalClient } = useClientsContext();
    const [isLoading, setIsLoading] = useState(false);

    const {
        handleSubmit,
        control,
        watch,
        setValue,
    } = useForm<TrackKeywordFormValues>({
        defaultValues: {
            keywords: project?.keywords.map((keyword) => keyword.name) ?? [],
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

    const addKeyword = useCallback((value: string) => {
        const trimmed = value.trim();
        if (!trimmed) return;

        const isDuplicate = keywords.some(
            (k) => k.toLowerCase() === trimmed.toLowerCase()
        );

        if (isDuplicate) {
            toast.error(`"${trimmed}" is already added`);
            return;
        }

        setValue("keywords", [...keywords, trimmed]);
    }, [keywords, setValue]);

    const handleDeleteKeyword = useCallback((index: number) => {
        const updated = keywords.filter((_, i) => i !== index);
        setValue("keywords", updated);
    }, [keywords, setValue]);

    const onSubmit = useCallback(async (data: TrackKeywordFormValues) => {
        if (!project) return;
        setIsLoading(true);

        try {
            await portalClient.createKeywords({ keywords: data.keywords });

            dispatch(setProject({ ...project, keywords: data.keywords }));
            dispatch(nextStep());
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    }, [dispatch, portalClient, project]);

    const handleBack = useCallback(() => dispatch(prevStep()), [dispatch]);

    const filteredSuggestions = listOfSuggestedKeywords.filter((suggestion) => !keywords.some((keyword) => keyword.toLowerCase() === suggestion.toLowerCase()));

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={2} mb={3} gap={4}>
                <Controller
                    name="newKeyword"
                    control={control}
                    render={({ field }) => (
                        <TextField
                            size="small"
                            fullWidth
                            label="Add Keyword"
                            variant="outlined"
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
                                        <IconButton onClick={handleAddKeyword} edge="end" size="small">
                                            <Plus size={16} />
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            }}
                        />
                    )}
                />

                <Box
                    sx={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 0.5,
                        px: 0.5,
                    }}
                >
                    {keywords.map((keyword, index) => (
                        <Chip
                            key={`${keyword}-${index}`}
                            label={keyword}
                            onDelete={() => handleDeleteKeyword(index)}
                            color="primary"
                            variant="outlined"
                            size="small"
                            sx={{ fontSize: "0.75rem", py: 0.5 }}
                        />
                    ))}
                </Box>

                {filteredSuggestions?.length > 0 && (
                    <Paper
                        variant="outlined"
                        sx={{
                            p: 1.5,
                            bgcolor: "background.default",
                            mt: 1,
                        }}
                    >
                        <Typography
                            variant="body2"
                            gutterBottom
                            sx={{ fontWeight: 500 }}
                        >
                            {`Suggested Keywords as per ${project?.name}`}
                        </Typography>
                        <Stack
                            direction="row"
                            spacing={0.5}
                            flexWrap="wrap"
                            gap={0.5}
                        >
                            {filteredSuggestions.map((suggestion, index) => (
                                <Chip
                                    key={index}
                                    label={suggestion}
                                    onClick={() => addKeyword(suggestion)}
                                    size="small"
                                    sx={{ fontSize: "0.7rem", py: 0.3 }}
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
                steps={steps}
                btnDisabled={isLoading || keywords.length === 0}
            />
        </form>
    );
}