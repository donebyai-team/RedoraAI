/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react/jsx-key */
"use client";

import {
    TextField,
    Typography,
    Chip,
    Box,
    Paper,
    Stack,
    Autocomplete,
} from "@mui/material";
import { X } from "lucide-react";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import {
    prevStep,
    setProject,
    setIsOnboardingDone
} from "../../../store/Onboarding/OnboardingSlice";
import React, { useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { useRouter } from "next/navigation";
import { routes } from "@doota/ui-core/routing";
import { Source } from "@doota/pb/doota/core/v1/core_pb";
import { useAuth } from "@doota/ui-core/hooks/useAuth";

interface SubredditFormValues {
    sources: Source[];
}

export default function SelectSourcesStep() {
    const dispatch = useDispatch();
    const routers = useRouter();
    const { user } = useAuth()
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const project = useAppSelector((state: RootState) => state.stepper.project);
    const listOfSuggestedSources = project?.suggestedSources ?? [];
    const { portalClient } = useClientsContext();

    const [loadingSubredditId, setLoadingSubredditId] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [inputValue, setInputValue] = useState("");
    const [pendingSources, setPendingSources] = useState<string[]>([]);

    const {
        handleSubmit,
        watch,
        setValue,
    } = useForm<SubredditFormValues>({
        defaultValues: {
            sources: project?.sources ?? [],
        },
    });

    const sources = watch("sources");

    const handleAddSubreddit = async (subredditName: string) => {
        const trimmed = subredditName.trim();
        if (!trimmed) return;

        const plainName = trimmed.replace(/^r\//i, "");
        const nameToSend = trimmed.startsWith("r/") ? trimmed : `r/${plainName}`;

        if (sources.some((s) => s.name.toLowerCase() === plainName.toLowerCase()) ||
            pendingSources.includes(plainName.toLowerCase())) {
            toast.error(`r/${plainName} is already being tracked`);
            return;
        }

        // Add to pending pills
        setPendingSources((prev) => [...prev, plainName.toLowerCase()]);
        setInputValue("");

        try {
            const result = await portalClient.addSource({ name: nameToSend });

            // const updatedProjects = await portalClient.self({});
            // const updatedSources = updatedProjects.project?.[0].sources.map(source => ({ id: source.id, name: source.name })) ?? [];
            const updatedSources = [...sources, result];

            setValue("sources", updatedSources);
            if (project) {
                dispatch(setProject({ ...project, sources: updatedSources }));
            }
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Failed to add";
            toast.error(message);
        } finally {
            // Remove from pending either way
            setPendingSources((prev) => prev.filter((name) => name !== plainName.toLowerCase()));
        }
    };

    const handleRemoveSubreddit = async (source: Source) => {
        setLoadingSubredditId(source.id);
        try {
            await portalClient.removeSource({ id: source.id });

            // 🔥 Get fresh list after remove
            // const updatedProjects = await portalClient.self({});
            // const updatedSources = updatedProjects.project?.[0].sources.map(source => ({ id: source.id, name: source.name })) ?? [];
            const updatedSources = sources.filter((item) => item.id !== source.id);

            setValue("sources", updatedSources);
            // toast.success(`Removed r/${source.name}`);
            if (project) {
                dispatch(setProject({ ...project, sources: updatedSources }));
            }

        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Failed to remove";
            toast.error(message);
        } finally {
            setLoadingSubredditId(null);
        }
    };

    const onSubmit = useCallback(
        async (data: SubredditFormValues) => {
            console.log(data);
            if (!project) return;
            setIsLoading(true);

            try {
                dispatch(setProject(project));
                dispatch(setIsOnboardingDone(true));
                if (user) {
                    user.isOnboardingDone = true;
                    user.projects = [...user.projects, project];
                }
                // toast.success("Sources saved successfully");
                routers.push(routes.app.home);
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong";
                toast.error(message);
            } finally {
                setIsLoading(false);
            }
        },
        [dispatch, portalClient, project, routers]  // Add router to dependencies
    );

    const handleBack = useCallback(() => dispatch(prevStep()), [dispatch]);

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            e.preventDefault();
            handleAddSubreddit(inputValue);
        }
    };

    const filteredSuggestions = listOfSuggestedSources.filter(
        (subreddit) => {
            const plainName = subreddit.replace(/^r\//i, "").toLowerCase();
            return !sources.some((s) => s.name.toLowerCase() === plainName);
        }
    );

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3} mb={5} gap={4}>
                <Autocomplete
                    multiple
                    freeSolo
                    options={[]}
                    value={sources.map((s) => s.name)}
                    inputValue={inputValue}
                    onInputChange={(_, newInputValue) => setInputValue(newInputValue)}
                    onChange={() => { }}
                    renderInput={(params) => (
                        <TextField
                            {...params}
                            label="Add Subreddit To Track"
                            placeholder="choose relevant subreddits to track keywords"
                            onKeyDown={handleKeyDown}
                            disabled={isLoading || loadingSubredditId !== null}
                        />
                    )}
                    renderTags={(_, getTagProps) => (
                        <>
                            {sources.map((option, index) => {
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
                                        sx={{ p: 2, fontSize: "0.75rem" }}
                                    />
                                );
                            })}
                            {pendingSources.map((name, index) => (
                                <Chip
                                    key={`pending-${index}`}
                                    label={`r/${name}`}
                                    color="default"
                                    disabled
                                    deleteIcon={
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
                                    }
                                    sx={{ p: 2, fontSize: "0.75rem", opacity: 0.6 }}
                                />
                            ))}
                        </>
                    )}

                />

                {filteredSuggestions.length > 0 && (
                    <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                        <Typography variant="subtitle2" gutterBottom>
                            {`Suggested subreddit as per ${project?.name}`}
                        </Typography>
                        <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                            {filteredSuggestions.map((subreddit) => {
                                const isLoading = loadingSubredditId === subreddit;

                                return (
                                    <Chip
                                        key={subreddit}
                                        label={subreddit}
                                        onClick={() => handleAddSubreddit(subreddit)}
                                        size="small"
                                        disabled={isLoading || loadingSubredditId !== null}
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
                )}

            </Stack>

            <StepperControls
                activeStep={activeStep}
                handleBack={handleBack}
                handleNext={handleSubmit(onSubmit)}
                steps={steps}
                btnDisabled={isLoading || sources.length === 0 || loadingSubredditId !== null}
            />
        </form>
    );
}
