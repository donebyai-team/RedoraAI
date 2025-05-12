"use client";

import {
    TextField,
    CircularProgress,
    Typography
} from "@mui/material";
import { Stack } from "@mui/system";
import StepperControls from "./StepperControls";
import { useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { steps } from "./MainForm";
import { useDispatch } from "react-redux";
import { nextStep, prevStep, resetStepper, setProjects } from "../../../store/Onboarding/OnboardingSlice";
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
};

const fields: FieldConfig[] = [
    {
        name: "website",
        label: "Product Website",
        placeholder: "https://example.com",
        rules: { required: "Product website is required" },
    },
    {
        name: "name",
        label: "Product Name",
        placeholder: "e.g., My Awesome Product",
        rules: { required: "Product name is required" },
    },
    {
        name: "description",
        label: "Product Description",
        placeholder: "Describe your product and its key features...",
        rules: { required: "Description is required" },
        multiline: true,
        rows: 3,
    },
    {
        name: "targetPersona",
        label: "Target Audience",
        placeholder: "e.g., Developers, Marketers, Small Business Owners",
        rules: { required: "Target audience is required" },
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
        if (!url.startsWith("http")) return;
        try {
            setLoadingMeta(true);
            const res = await fetch(`/api/fetch-meta?url=${encodeURIComponent(url)}`);
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
                suggestedKeywords: result.suggestedKeywords ?? [],
                suggestedSources: result.suggestedSources ?? [],
            };
            dispatch(setProjects(newPayload));

            toast.success("Product Information saved successfully");
            dispatch(nextStep());
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    };

    const handleBack = () => dispatch(prevStep());
    const handleReset = () => dispatch(resetStepper());

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3} mb={5}>
                {fields.map((field) => (
                    <Controller
                        key={field.name}
                        name={field.name as keyof ProductFormValues}
                        control={control}
                        rules={field.rules}
                        render={({ field: controllerField }) => (
                            <TextField
                                {...controllerField}
                                fullWidth
                                label={field.label}
                                placeholder={field.placeholder}
                                error={!!errors[field.name as keyof ProductFormValues]}
                                helperText={errors[field.name as keyof ProductFormValues]?.message}
                                disabled={isLoading}
                                multiline={field.multiline}
                                rows={field.rows}
                            />
                        )}
                    />
                ))}

                {loadingMeta && (
                    <Typography variant="body2" color="text.secondary" sx={{ display: "flex", alignItems: "center" }}>
                        <CircularProgress size={14} sx={{ mr: 1 }} />
                        Fetching site metadata...
                    </Typography>
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