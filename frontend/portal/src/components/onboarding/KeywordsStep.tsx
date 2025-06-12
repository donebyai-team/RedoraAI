
import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X, Plus, Lightbulb } from "lucide-react";
import { toast } from "@/hooks/use-toast";
import { useForm } from "react-hook-form";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { portalClient } from "@/services/grpc";
import { nextStep, prevStep, setProject } from "@/store/Onboarding/OnboardingSlice";
import { useAuth } from "@doota/ui-core/hooks/useAuth";

export interface TrackKeywordFormValues {
  keywords: string[];
  newKeyword: string;
}

export default function KeywordsStep() {

  const { planDetails } = useAuth();
  const MAX_KEYWORDS = planDetails.subscription?.maxKeywords as number;
  const project = useAppSelector((state) => state.stepper.project);
  const [suggestionLoading, setSuggestionLoading] = useState<boolean>(true);
  const [isLoading, setIsLoading] = useState(false);
  const listOfSuggestedKeywords = project?.suggestedKeywords ?? [];
  const dispatch = useAppDispatch();

  const {
    handleSubmit,
    watch,
    setValue,
    register,
    formState: { errors }
  } = useForm<TrackKeywordFormValues>({
    defaultValues: {
      keywords: project?.keywords.map((keyword) => keyword.name) ?? [],
      newKeyword: "",
    },
  });

  const keywords = watch("keywords");
  const newKeyword = watch("newKeyword");

  const handleAddKeyword = () => {
    const trimmed = newKeyword.trim();

    if (!trimmed) return;

    const isDuplicate = keywords.some(
      (k) => k.toLowerCase() === trimmed.toLowerCase()
    );

    if (isDuplicate) {
      toast({
        title: "Error",
        description: `"${trimmed}" is already added`,
        variant: "destructive"
      });
      return;
    }

    setValue("keywords", [...keywords, trimmed]);
    setValue("newKeyword", "");
  };

  const addSuggestedKeyword = (value: string) => {
    const trimmed = value.trim();
    if (!trimmed) return;

    const isDuplicate = keywords.some(
      (k) => k.toLowerCase() === trimmed.toLowerCase()
    );

    if (isDuplicate) {
      toast({
        title: "Error",
        description: `"${trimmed}" is already added`,
        variant: "destructive"
      });
      return;
    }

    setValue("keywords", [...keywords, trimmed]);
  };

  const removeKeyword = (index: number) => {
    const updated = keywords.filter((_, i) => i !== index);
    setValue("keywords", updated);
  };

  const onSubmit = async () => {
    if (!project) return;
    setIsLoading(true);

    try {
      const result = await portalClient.createKeywords({ keywords });

      dispatch(setProject({ ...project, keywords: result.keywords }));
      dispatch(nextStep());
    } catch (err: any) {
      // console.log("###_eerr ", err);
      const message = err?.response?.data?.message || err.message || "Something went wrong";
      toast({
        title: "Error",
        description: message,
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!project) return;

    const fetchSuggestions = async () => {
      setSuggestionLoading(true);
      try {
        const result = await portalClient.suggestKeywordsAndSources({});
        dispatch(setProject(result));
      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Failed to fetch suggestions";
        toast({
          title: "Error",
          description: message,
          variant: "destructive"
        });
      } finally {
        setSuggestionLoading(false);
      }
    };

    fetchSuggestions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const availableSuggestions = listOfSuggestedKeywords.filter((suggestion) => !keywords.some((keyword) => keyword.toLowerCase() === suggestion.toLowerCase()));

  return (
    <div className="space-y-6">
      {/* Current Keywords */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label>Your Keywords ({keywords.length}/{MAX_KEYWORDS})</Label>
          <Badge variant="secondary">{MAX_KEYWORDS - keywords.length} remaining</Badge>
        </div>

        {keywords.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {keywords.map((keyword, index) => (
              <Badge key={keyword} variant="default" className="flex items-center gap-1">
                {keyword}
                <button
                  onClick={() => removeKeyword(index)}
                  className="ml-1 hover:bg-primary-foreground/20 rounded-full p-0.5"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : (
          <p className="text-muted-foreground text-sm">No keywords added yet</p>
        )}
      </div>

      {/* Add New Keyword */}
      <form onSubmit={handleSubmit(handleAddKeyword)}>
        <div className="space-y-2">
          <Label htmlFor="newKeyword">Add Keyword</Label>
          <div className="flex gap-2">
            <Input
              id="newKeyword"
              {...register("newKeyword", {
                validate: (value) => {
                  const trimmedKeyword = value.trim();

                  if (!trimmedKeyword) {
                    return "Please enter a keyword";
                  }

                  if (keywords.includes(trimmedKeyword)) {
                    return "This keyword is already added";
                  }

                  if (keywords.length >= MAX_KEYWORDS) {
                    return `You can only add up to ${MAX_KEYWORDS} keywords`;
                  }

                  return true;
                },
              })}
              placeholder="Enter a keyword to track"
              className={errors.newKeyword?.message ? "border-destructive" : ""}
              disabled={keywords.length >= MAX_KEYWORDS}
            />
            <Button
              type="submit"
              disabled={keywords.length >= MAX_KEYWORDS}
              size="icon"
            >
              <Plus className="w-4 h-4" />
            </Button>
          </div>
          {errors.newKeyword?.message && <p className="text-sm text-destructive">{errors.newKeyword?.message}</p>}
        </div>
      </form>

      {/* Suggested Keywords */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center gap-2">
            <Lightbulb className="w-4 h-4" />
            Suggested Keywords
          </CardTitle>
        </CardHeader>
        <CardContent>
          {suggestionLoading ? (
            <div className="flex justify-center items-center my-14">
              <p className="text-sm text-muted-foreground">
                {`Please wait while we are generating suggestions keywords or subreddit for you...`}
              </p>
            </div>
          ) : availableSuggestions.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {availableSuggestions.map((suggestion) => (
                <Button
                  key={suggestion}
                  variant="outline"
                  size="sm"
                  onClick={() => addSuggestedKeyword(suggestion)}
                  disabled={keywords.length >= MAX_KEYWORDS}
                  className="h-8"
                >
                  <Plus className="w-3 h-3 mr-1" />
                  {suggestion}
                </Button>
              ))}
            </div>
          ) : (
            <div className="flex justify-center items-center my-14">
              <p className="text-sm text-muted-foreground">
                {`No suggestion keywords found.`}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex justify-between">
        <Button variant="outline" onClick={() => dispatch(prevStep())} disabled={isLoading}>
          Back to Product Details
        </Button>
        <Button onClick={onSubmit} disabled={isLoading || keywords.length === 0}>
          Continue to Subreddits
        </Button>
      </div>
    </div>
  );
}
