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
import { nextStep, prevStep, resetStepper } from "../../../store/Onboarding/OnboardingSlice";
import { useEffect, useState } from "react";
import { useForm, Controller } from "react-hook-form";

interface ProductFormValues {
    website: string;
    name: string;
    description: string;
    targetPersona: string;
}

export default function ProductInformationStep() {
    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
        watch,
    } = useForm<ProductFormValues>({
        defaultValues: {
            website: "",
            name: "",
            description: "",
            targetPersona: "",
        },
    });

    const website = watch("website");
    const [loadingMeta, setLoadingMeta] = useState(false);

    // 🔁 Debounce metadata fetch
    useEffect(() => {
        if (!website) return;

        const timer = setTimeout(() => {
            if (!website.startsWith("http")) return;

            setLoadingMeta(true);
            fetch(`/api/fetch-meta?url=${encodeURIComponent(website)}`)
                .then(res => res.json())
                .then(data => {
                    setValue("name", data.title || "");
                    setValue("description", data.description || "");
                })
                .catch(() => { })
                .finally(() => setLoadingMeta(false));
        }, 700); // 700ms debounce

        return () => clearTimeout(timer);
    }, [website, setValue]);

    const onSubmit = (data: ProductFormValues) => {
        // You can post `data` here if this is the final step
        console.log("###_debug_data ", data);
        dispatch(nextStep());
    };

    const handleBack = () => {
        dispatch(prevStep());
    };

    const handleReset = () => {
        dispatch(resetStepper());
    };

    return (<>
        <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={3} mb={5}>
                <Controller
                    name="website"
                    control={control}
                    rules={{ required: "Product website is required" }}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            fullWidth
                            label="Product Website"
                            placeholder="https://example.com"
                            error={!!errors.website}
                            helperText={errors.website?.message}
                        />
                    )}
                />

                {loadingMeta && (
                    <Typography variant="body2" color="text.secondary">
                        <CircularProgress size={14} sx={{ mr: 1 }} />
                        Fetching site metadata...
                    </Typography>
                )}

                <Controller
                    name="name"
                    control={control}
                    rules={{ required: "Product name is required" }}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            fullWidth
                            label="Product Name"
                            placeholder="e.g., My Awesome Product"
                            error={!!errors.name}
                            helperText={errors.name?.message}
                        />
                    )}
                />

                <Controller
                    name="description"
                    control={control}
                    rules={{ required: "Description is required" }}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            fullWidth
                            multiline
                            rows={3}
                            label="Product Description"
                            placeholder="Describe your product and its key features..."
                            error={!!errors.description}
                            helperText={errors.description?.message}
                        />
                    )}
                />

                <Controller
                    name="targetPersona"
                    control={control}
                    rules={{ required: "Target audience is required" }}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            fullWidth
                            label="Target Audience"
                            placeholder="e.g., Developers, Marketers, Small Business Owners"
                            error={!!errors.targetPersona}
                            helperText={errors.targetPersona?.message}
                        />
                    )}
                />
            </Stack>

            <StepperControls
                activeStep={activeStep}
                handleBack={handleBack}
                handleNext={handleSubmit(onSubmit)}
                handleReset={handleReset}
                steps={steps}
            />
        </form>
    </>);
}