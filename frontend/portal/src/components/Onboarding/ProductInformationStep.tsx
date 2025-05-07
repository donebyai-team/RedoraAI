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
// import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";

interface ProductFormValues {
    website: string;
    name: string;
    description: string;
    targetPersona: string;
}

export default function ProductInformationStep() {
    const dispatch = useDispatch();
    const activeStep = useAppSelector((state: RootState) => state.stepper.activeStep);
    // const { portalClient } = useClientsContext()
    const {
        control,
        handleSubmit,
        setValue,
        clearErrors,
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
    const [isLoading, setIsLoading] = useState(false)

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

                    // ✅ Clear errors only if value is not empty
                    if (data.title) {
                        clearErrors("name");
                    }
                    if (data.description) {
                        clearErrors("description");
                    }
                })
                .catch(() => { })
                .finally(() => setLoadingMeta(false));
        }, 700); // 700ms debounce

        return () => clearTimeout(timer);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [website, setValue]);

    const onSubmit = async (data: ProductFormValues) => {
        // You can post `data` here if this is the final step
        console.log("###_debug_data ", data);

        setIsLoading(true)

        try {
            // await portalClient.createOrEditProject({  })
            await new Promise(resolve => setTimeout(resolve, 2300));

            const msg = "Product Information saved successfully";
            toast.success(msg)
            dispatch(nextStep());
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong"
            toast.error(message)
        } finally {
            setIsLoading(false)
        }
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
                            disabled={isLoading}
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
                            disabled={isLoading}
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
                            disabled={isLoading}
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
                            disabled={isLoading}
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
                btnDisabled={isLoading}
            />
        </form>
    </>);
}