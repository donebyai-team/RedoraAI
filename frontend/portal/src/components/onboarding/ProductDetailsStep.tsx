import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { toast } from "@/hooks/use-toast";
import { Globe, Loader2 } from "lucide-react";

interface ProductDetails {
  website: string;
  productName: string;
  description: string;
  targetPersona: string;
}

interface ProductDetailsStepProps {
  data: ProductDetails;
  onUpdate: (data: ProductDetails) => void;
  onNext: () => void;
}

const fetchWebsiteMetadata = async (url: string) => {
  try {
    // In a real app, you'd call your backend API or use a service like LinkPreview
    // For now, we'll simulate the API call
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Mock response based on common website patterns
    const domain = new URL(url).hostname.replace('www.', '');
    const companyName = domain.split('.')[0];

    return {
      title: `${companyName.charAt(0).toUpperCase() + companyName.slice(1)}`,
      description: `Innovative solutions and services provided by ${companyName}`,
    };
  } catch (error) {
    throw new Error("Failed to fetch website metadata");
  }
};

export default function ProductDetailsStep({ data, onUpdate, onNext }: ProductDetailsStepProps) {
  const [formData, setFormData] = useState<ProductDetails>(data);
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState<Partial<ProductDetails>>({});

  useEffect(() => {
    onUpdate(formData);
  }, [formData, onUpdate]);

  const handleWebsiteChange = (website: string) => {
    setFormData(prev => ({ ...prev, website }));
    setErrors(prev => ({ ...prev, website: "" }));
  };

  const fetchMetadata = async () => {
    if (!formData.website) {
      setErrors(prev => ({ ...prev, website: "Please enter a website URL" }));
      return;
    }

    try {
      new URL(formData.website);
    } catch {
      setErrors(prev => ({ ...prev, website: "Please enter a valid URL" }));
      return;
    }

    setIsLoading(true);
    try {
      const metadata = await fetchWebsiteMetadata(formData.website);
      setFormData(prev => ({
        ...prev,
        productName: metadata.title,
        description: metadata.description,
      }));
      toast({
        title: "Website metadata fetched!",
        description: "Product name and description have been pre-filled.",
      });
    } catch (error) {
      toast({
        title: "Failed to fetch metadata",
        description: "Please fill in the details manually.",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const validateForm = () => {
    const newErrors: Partial<ProductDetails> = {};

    if (!formData.website.trim()) {
      newErrors.website = "Website is required";
    } else {
      try {
        new URL(formData.website);
      } catch {
        newErrors.website = "Please enter a valid URL";
      }
    }

    if (!formData.productName.trim()) {
      newErrors.productName = "Product name is required";
    } else if (formData.productName.length < 3 || formData.productName.length > 30) {
      newErrors.productName = "Product name must be between 3 and 30 characters";
    } else {
      const wordCount = formData.productName.trim().split(/\s+/).length;
      if (wordCount > 3) {
        newErrors.productName = "Product name can have maximum 3 words";
      }
    }

    if (!formData.description.trim()) {
      newErrors.description = "Description is required";
    } else if (formData.description.length < 10) {
      newErrors.description = "Description must be at least 10 characters";
    }

    if (!formData.targetPersona.trim()) {
      newErrors.targetPersona = "Target persona is required";
    } else if (formData.targetPersona.length < 10) {
      newErrors.targetPersona = "Target persona must be at least 10 characters";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = () => {
    if (validateForm()) {
      onNext();
    }
  };

  return (
    <div className="space-y-6">
      {/* Website URL */}
      <div className="space-y-2">
        <Label htmlFor="website">Website URL *</Label>
        <div className="flex gap-2">
          <div className="flex-1">
            <Input
              id="website"
              type="url"
              placeholder="https://example.com - We'll automatically fetch your product details from your website to save you time"
              value={formData.website}
              onChange={(e) => handleWebsiteChange(e.target.value)}
              className={errors.website ? "border-destructive" : ""}
            />
            {errors.website && (
              <p className="text-sm text-destructive mt-1">{errors.website}</p>
            )}
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={fetchMetadata}
            disabled={isLoading || !formData.website}
          >
            {isLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Globe className="w-4 h-4" />
            )}
            Fetch
          </Button>
        </div>
      </div>

      {/* Product Name */}
      <div className="space-y-2">
        <Label htmlFor="productName">Product Name *</Label>
        <Input
          id="productName"
          value={formData.productName}
          onChange={(e) => setFormData(prev => ({ ...prev, productName: e.target.value }))}
          placeholder="Keep it simple and recognizable - we'll use this to identify when people mention your product"
          className={errors.productName ? "border-destructive" : ""}
        />
        {errors.productName && (
          <p className="text-sm text-destructive">{errors.productName}</p>
        )}
        <p className="text-xs text-muted-foreground">
          {formData.productName.length}/30 characters • {formData.productName.trim().split(/\s+/).filter(word => word.length > 0).length}/3 words
        </p>
      </div>

      {/* Description */}
      <div className="space-y-2">
        <Label htmlFor="description">Product Description *</Label>
        <Textarea
          id="description"
          value={formData.description}
          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
          placeholder="Describe what your product does and what problems it solves - this helps us find relevant discussions where people might need your solution"
          className={errors.description ? "border-destructive" : ""}
          rows={3}
        />
        {errors.description && (
          <p className="text-sm text-destructive">{errors.description}</p>
        )}
        <p className="text-xs text-muted-foreground">
          {formData.description.length} characters (minimum 10)
        </p>
      </div>

      {/* Target Persona */}
      <div className="space-y-2">
        <Label htmlFor="targetPersona">Target Persona *</Label>
        <Input
          id="targetPersona"
          value={formData.targetPersona}
          onChange={(e) => setFormData(prev => ({ ...prev, targetPersona: e.target.value }))}
          placeholder="Who is your ideal customer? This helps us identify the right communities and conversations where your audience hangs out"
          className={errors.targetPersona ? "border-destructive" : ""}
        />
        {errors.targetPersona && (
          <p className="text-sm text-destructive">{errors.targetPersona}</p>
        )}
        <p className="text-xs text-muted-foreground">
          {formData.targetPersona.length} characters (minimum 10)
        </p>
      </div>

      <div className="flex justify-end">
        <Button onClick={handleNext} disabled={isLoading}>
          Continue to Keywords
        </Button>
      </div>
    </div>
  );
}
