"use client";

import {
    TextField,
    CircularProgress,
    Typography,
} from "@mui/material";
import { Stack } from "@mui/system";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import { nextStep, prevStep, setProjects } from "../../../store/Onboarding/OnboardingSlice";
import { useEffect, useState, useCallback } from "react";
import { useForm, Controller } from "react-hook-form";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";

interface ProductFormValues {
    website: string;
    name: string;
    description: string;
    targetPersona: string;
}

type FieldConfig = {
    name: keyof ProductFormValues;
    label: string;
    placeholder: string;
    rules: Record<string, any>;
    multiline?: boolean;
    rows?: number;
    helperText?: string;
};

const fields: FieldConfig[] = [
    {
        name: "website",
        label: "Website URL",
        placeholder: "donebyai.team",
        rules: { required: "Website url is required" },
    },
    {
        name: "name",
        label: "Product Name",
        placeholder: "e.g., MiraAI",
        rules: { required: "Product name is required" },
    },
    {
        name: "description",
        label: "Product Description",
        placeholder: "MiraAI talks to your inbound leads over SMS and WhatsApp like a human, qualifying prospects and scheduling calls with your sales team.",
        rules: { required: "Description is required" },
        multiline: true,
        rows: 3,
        helperText: "Add what your product does in brief between 15-20 words",
    },
    {
        name: "targetPersona",
        label: "Target Audience",
        placeholder: "e.g., Developers, Marketers, Small Business Owners",
        rules: { required: "Target audience is required" },
        multiline: true,
        rows: 3,
        helperText: "Briefly explain what kind of customer you are targeting in 15-20 words",
    },
];

export default function ProductInformationStep() {
    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const projects = useAppSelector((state: RootState) => state.stepper.projects);
    const { portalClient } = useClientsContext();

    const {
        control,
        handleSubmit,
        setValue,
        clearErrors,
        formState: { errors },
        watch,
    } = useForm<ProductFormValues>({
        defaultValues: {
            website: projects?.website ?? "",
            name: projects?.name ?? "",
            description: projects?.description ?? "",
            targetPersona: projects?.targetPersona ?? "",
        },
    });

    const website = watch("website");
    const [loadingMeta, setLoadingMeta] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    const fetchMeta = useCallback(async (url: string) => {
        if (!url) return;

        // Prepend https:// if no scheme is present
        const normalizedUrl = url.startsWith("http://") || url.startsWith("https://")
            ? url
            : `https://${url}`;

        try {
            setLoadingMeta(true);
            const res = await fetch(`/api/fetch-meta?url=${encodeURIComponent(normalizedUrl)}`);
            const data = await res.json();

            setValue("name", data.title || "");
            setValue("description", data.description || "");

            if (data.title) clearErrors("name");
            if (data.description) clearErrors("description");
        } catch {
            // silently fail
        } finally {
            setLoadingMeta(false);
        }
    }, [setValue, clearErrors]);

    useEffect(() => {
        const timer = setTimeout(() => {
            if (website && website !== projects?.website) {
                fetchMeta(website);
            }
        }, 700);

        return () => clearTimeout(timer);
    }, [website, projects?.website, fetchMeta]);

    const onSubmit = async (data: ProductFormValues) => {
        setIsLoading(true);

        try {
            const body = {
                ...(projects?.id && { id: projects.id }),
                ...data,
            };

            const result = await portalClient.createOrEditProject(body);
            const newPayload = {
                ...projects,
                id: result.id ?? projects?.id,
                name: result.name,
                description: result.description,
                website: result.website,
                targetPersona: result.targetPersona,
                suggestedKeywords: result?.suggestedKeywords ?? [],
                suggestedSources: result?.suggestedSources ?? [],
                // suggestedKeywords: ["SEO AI", "DEV"],
                // suggestedSources: ["r/php", "r/java", "r/laravel"],
            };
            dispatch(setProjects(newPayload));
            dispatch(nextStep());
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    };

    const handleBack = () => dispatch(prevStep());

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={2} mb={3} gap={3.8}>
                {fields.map((field) => (
                    <Controller
                        key={field.name}
                        name={field.name}
                        control={control}
                        rules={field.rules}
                        render={({ field: controllerField }) => (
                            <TextField
                                {...controllerField}
                                size="small"
                                fullWidth
                                label={field.label}
                                placeholder={field.placeholder}
                                multiline={field.multiline}
                                rows={field.rows}
                                disabled={isLoading || loadingMeta}
                                error={!!errors[field.name]}
                                helperText={errors[field.name]?.message ?? field.helperText}
                                FormHelperTextProps={{
                                    sx: { ml: 1.5, fontSize: "0.75rem" },
                                }}
                            />
                        )}
                    />
                ))}

                {loadingMeta && (
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{ display: "flex", alignItems: "center", gap: 1 }}
                    >
                        <CircularProgress size={12} />
                        Fetching site metadata...
                    </Typography>
                )}
            </Stack>

            <StepperControls
                activeStep={activeStep}
                handleBack={handleBack}
                handleNext={handleSubmit(onSubmit)}
                steps={steps}
                btnDisabled={isLoading || loadingMeta}
            />
        </form>
    );
}