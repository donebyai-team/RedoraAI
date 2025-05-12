/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react/jsx-key */
"use client";

import {
    TextField,
    Typography,
    Chip,
    Box,
    InputAdornment,
    Paper,
    Stack,
    Autocomplete,
} from "@mui/material";
import { Search, X } from "lucide-react";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import {
    prevStep,
    setProjects,
    SourcesTypes,
} from "../../../store/Onboarding/OnboardingSlice";
import React, { useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { useRouter } from "next/navigation";
import { routes } from "@doota/ui-core/routing";

interface SubredditFormValues {
    sources: SourcesTypes[];
}

export default function SelectSourcesStep() {
    const dispatch = useDispatch();
    const routers = useRouter();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const projects = useAppSelector((state: RootState) => state.stepper.projects);
    const listOfSuggestedSources = projects?.suggestedSources ?? [];
    const { portalClient } = useClientsContext();

    const [loadingSubredditId, setLoadingSubredditId] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [inputValue, setInputValue] = useState("");

    const {
        handleSubmit,
        watch,
        setValue,
    } = useForm<SubredditFormValues>({
        defaultValues: {
            sources: projects?.sources ?? [],
        },
    });

    const sources = watch("sources");

    const handleAddSubreddit = async (subredditName: string) => {
        const trimmed = subredditName.trim();

        if (!trimmed) return;

        // 🔎 Remove r/ if present for checking
        const plainName = trimmed.replace(/^r\//i, "");

        // Check duplicate in sources (compare name only)
        if (sources.some((s) => s.name.toLowerCase() === plainName.toLowerCase())) {
            toast.error(`r/${plainName} is already added`);
            return;
        }

        // 🔥 Ensure the name we send has r/ prefix
        const nameToSend = trimmed.startsWith("r/") ? trimmed : `r/${plainName}`;

        setLoadingSubredditId(plainName);  // use plain name for loading check
        try {
            await portalClient.addSource({ name: nameToSend });

            // 🔄 Get fresh list after add
            const updatedProjects = await portalClient.getProjects({});
            const updatedSources = updatedProjects.projects?.[0].sources.map(source => ({ id: source.id, name: source.name })) ?? [];

            setValue("sources", updatedSources);
            toast.success(`Added ${nameToSend}`);
            setInputValue("");

        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Failed to add";
            toast.error(message);
        } finally {
            setLoadingSubredditId(null);
        }
    };

    const handleRemoveSubreddit = async (source: SourcesTypes) => {
        setLoadingSubredditId(source.id);
        try {
            await portalClient.removeSource({ id: source.id });

            // 🔥 Get fresh list after remove
            const updatedProjects = await portalClient.getProjects({});
            const updatedSources = updatedProjects.projects?.[0].sources.map(source => ({ id: source.id, name: source.name })) ?? [];

            setValue("sources", updatedSources);
            toast.success(`Removed r/${source.name}`);

        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Failed to remove";
            toast.error(message);
        } finally {
            setLoadingSubredditId(null);
        }
    };

    const onSubmit = useCallback(
        async (data: SubredditFormValues) => {
            if (!projects) return;
            setIsLoading(true);

            try {
                dispatch(setProjects({ ...projects, sources: data.sources }));
                toast.success("Sources saved successfully");
                routers.push(routes.app.home);
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong";
                toast.error(message);
            } finally {
                setIsLoading(false);
            }
        },
        [dispatch, portalClient, projects, routers]  // Add router to dependencies
    );

    const handleBack = useCallback(() => dispatch(prevStep()), [dispatch]);

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            e.preventDefault();
            handleAddSubreddit(inputValue);
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3} mb={5}>
                <Autocomplete
                    multiple
                    freeSolo
                    options={listOfSuggestedSources}
                    value={sources.map((s) => s.name)}
                    inputValue={inputValue}
                    onInputChange={(_, newInputValue) => setInputValue(newInputValue)}
                    onChange={() => { }}
                    renderInput={(params) => (
                        <TextField
                            {...params}
                            label="Search Subreddits"
                            placeholder="Type subreddit and press Enter"
                            InputProps={{
                                ...params.InputProps,
                                startAdornment: (
                                    <>
                                        <InputAdornment position="start">
                                            <Search size={20} />
                                        </InputAdornment>
                                        {params.InputProps.startAdornment}
                                    </>
                                ),
                            }}
                            onKeyDown={handleKeyDown}
                        />
                    )}
                    renderTags={(_, getTagProps) =>
                        sources.map((option, index) => {
                            const isLoading = loadingSubredditId === option.id;
                            return (
                                <Chip
                                    label={`r/${option.name}`}
                                    {...getTagProps({ index })}
                                    color="primary"
                                    onDelete={() => handleRemoveSubreddit(option)}
                                    deleteIcon={
                                        isLoading ? (
                                            <Box
                                                component="span"
                                                sx={{
                                                    border: '2px solid transparent',
                                                    borderTop: '2px solid currentColor',
                                                    borderRadius: '50%',
                                                    width: 16,
                                                    height: 16,
                                                    animation: 'spin 0.6s linear infinite',
                                                }}
                                            />
                                        ) : (
                                            <X size={16} />
                                        )
                                    }
                                    disabled={isLoading}
                                />
                            );
                        })
                    }
                />

                <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                    <Typography variant="subtitle2" gutterBottom>
                        Popular Subreddits
                    </Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                        {listOfSuggestedSources.map((subreddit) => {
                            const isSelected = sources.some((s) => s.name === subreddit);
                            const isLoading = loadingSubredditId === subreddit;

                            return (
                                <Chip
                                    key={subreddit}
                                    label={`${subreddit}`}
                                    onClick={() => handleAddSubreddit(subreddit)}
                                    size="small"
                                    color={isSelected ? "primary" : "default"}
                                    disabled={isLoading || isSelected}
                                    deleteIcon={
                                        isLoading ? (
                                            <Box
                                                component="span"
                                                sx={{
                                                    border: '2px solid transparent',
                                                    borderTop: '2px solid currentColor',
                                                    borderRadius: '50%',
                                                    width: 16,
                                                    height: 16,
                                                    animation: 'spin 0.6s linear infinite',
                                                }}
                                            />
                                        ) : undefined
                                    }
                                />
                            );
                        })}
                    </Stack>
                </Paper>
            </Stack>

            <StepperControls
                activeStep={activeStep}
                handleBack={handleBack}
                handleNext={handleSubmit(onSubmit)}
                steps={steps}
                btnDisabled={isLoading || sources.length === 0} // Disable next if no sources are selected
            />
        </form>
    );
}
