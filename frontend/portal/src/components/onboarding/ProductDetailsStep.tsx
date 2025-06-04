import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { toast } from "@/hooks/use-toast";
// import { Globe, Loader2 } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { useForm } from "react-hook-form";
import { nextStep, setProject } from "@/store/Onboarding/OnboardingSlice";
import { portalClient } from "@/services/grpc";
import { setLoading } from "@/store/Source/sourceSlice";

interface ProductFormValues {
  website: string;
  name: string;
  description: string;
  targetPersona: string;
}

export default function ProductDetailsStep() {

  const project = useAppSelector((state) => state.stepper.project);
  const dispatch = useAppDispatch();
  const [isLoading, setIsLoading] = useState(false);

  const {
    handleSubmit,
    setValue,
    clearErrors,
    register,
    formState: { errors },
    watch,
  } = useForm<ProductFormValues>({
    defaultValues: {
      website: project?.website ?? "",
      name: project?.name ?? "",
      description: project?.description ?? "",
      targetPersona: project?.targetPersona ?? "",
    },
  });

  const website = watch("website");
  const name = watch("name");
  const description = watch("description");
  const targetPersona = watch("targetPersona");

  const fetchMeta = useCallback(async (url: string) => {
    if (!url) return;

    const normalizedUrl = url.startsWith("http://") || url.startsWith("https://")
      ? url
      : `https://${url}`;

    try {
      setLoading(true);
      const res = await fetch(`/api/fetch-meta?url=${encodeURIComponent(normalizedUrl)}`);
      const data = await res.json();

      setValue("name", data.title || "");
      setValue("description", data.description || "");

      if (data.title) {
        clearErrors("name");
      }

      if (data.description) {
        clearErrors("description");
      }
    } catch (err) {
      console.log(err);
      setValue("name", "");
      setValue("description", "");
      clearErrors("name");
      clearErrors("description");
    } finally {
      setLoading(false);
    }
  }, [setValue, clearErrors]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (website && website !== project?.website) {
        fetchMeta(website);
      }
    }, 700);

    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [website]);

  const onSubmit = async (data: ProductFormValues) => {
    const cleanedData = Object.fromEntries(Object.entries(data).map(([k, v]) => [k, typeof v === "string" ? v.trim() : v])) as ProductFormValues;

    setIsLoading(true);

    try {
      const body = {
        ...(project?.id && { id: project.id }),
        ...cleanedData,
      };

      const result = await portalClient.createOrEditProject(body);
      dispatch(setProject(result));
      dispatch(nextStep());
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="space-y-6">

        {/* Website URL */}
        <div className="space-y-2">
          <Label htmlFor="website">Website URL *</Label>
          <div className="flex gap-2">
            <div className="flex-1">
              <Input
                id="website"                
                placeholder="https://example.com - We'll automatically fetch your product details from your website to save you time"
                {...register('website', {
                  required: "Website url is required",
                  validate: (value: string) => value.trim().length > 0 || "This field cannot be empty.",
                })}
                className={errors.website?.message ? "border-destructive" : ""}
                disabled={isLoading}
              />
              {errors.website?.message && (
                <p className="text-sm text-destructive mt-1">{errors.website?.message}</p>
              )}
            </div>
          </div>
        </div>

        {/* Product Name */}
        <div className="space-y-2">
          <Label htmlFor="productName">Product Name *</Label>
          <Input
            id="productName"
            {...register('name', {
              required: "Product name is required",
              validate: (value) => {
                const trimmed = value.trim();

                if (!trimmed) {
                  return "Product name is required";
                }

                if (trimmed.length < 3 || trimmed.length > 30) {
                  return "Product name must be between 3 and 30 characters";
                }

                const wordCount = trimmed.split(/\s+/).length;
                if (wordCount > 3) {
                  return "Product name can have maximum 3 words";
                }

                return true;
              },
            })}
            placeholder="Keep it simple and recognizable - we'll use this to identify when people mention your product"
            className={errors.name?.message ? "border-destructive" : ""}
            disabled={isLoading}
          />
          {errors.name?.message && (
            <p className="text-sm text-destructive">{errors.name?.message}</p>
          )}
          <p className="text-xs text-muted-foreground">
            {name.length}/30 characters • {name.trim().split(/\s+/).filter(word => word.length > 0).length}/3 words
          </p>
        </div>

        {/* Description */}
        <div className="space-y-2">
          <Label htmlFor="description">Product Description *</Label>
          <Textarea
            id="description"
            {...register("description", {
              required: "Description is required",
              validate: (value) => {
                const trimmed = value.trim();

                if (!trimmed) return "Description is required";
                if (trimmed.length < 10) return "Description must be at least 10 characters";

                return true;
              },
            })}
            placeholder="Describe what your product does and what problems it solves - this helps us find relevant discussions where people might need your solution"
            className={errors.description?.message ? "border-destructive" : ""}
            rows={3}
            disabled={isLoading}
          />
          {errors.description?.message && (
            <p className="text-sm text-destructive">{errors.description?.message}</p>
          )}
          <p className="text-xs text-muted-foreground">
            {description.length} characters (minimum 10)
          </p>
        </div>

        {/* Target Persona */}
        <div className="space-y-2">
          <Label htmlFor="targetPersona">Target Persona *</Label>
          <Input
            id="targetPersona"
            {...register("targetPersona", {
              required: "Target persona is required",
              validate: (value) => {
                const trimmed = value.trim();

                if (!trimmed) return "Target persona is required";
                if (trimmed.length < 10) return "Target persona must be at least 10 characters";

                return true;
              },
            })}
            placeholder="Who is your ideal customer? This helps us identify the right communities and conversations where your audience hangs out"
            className={errors.targetPersona?.message ? "border-destructive" : ""}
            disabled={isLoading}
          />
          {errors.targetPersona?.message && (
            <p className="text-sm text-destructive">{errors.targetPersona?.message}</p>
          )}
          <p className="text-xs text-muted-foreground">
            {targetPersona.length} characters (minimum 10)
          </p>
        </div>

        <div className="flex justify-end">
          <Button type="submit" disabled={isLoading}>
            Continue to Keywords
          </Button>
        </div>
      </div>
    </form>
  );
}
