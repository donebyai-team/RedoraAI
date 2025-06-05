
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X, Plus, Lightbulb } from "lucide-react";
import { useForm } from "react-hook-form";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { useRouter } from "next/navigation";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { finishOnboarding, prevStep, setIsOnboardingDone, setProject } from "@/store/Onboarding/OnboardingSlice";
import { toast } from "@/hooks/use-toast";
import { portalClient } from "@/services/grpc";
import { Source } from "@doota/pb/doota/core/v1/core_pb";
import { routes } from "@doota/ui-core/routing";

interface SubredditFormValues {
  sources: Source[];
  newSubreddit: string;
}

export default function SubredditsStep() {

  const routers = useRouter();
  const { setUser, planDetails } = useAuth()
  const project = useAppSelector((state) => state.stepper.project);
  const listOfSuggestedSources = project?.suggestedSources ?? [];
  const [loadingSubredditId, setLoadingSubredditId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [pendingSources, setPendingSources] = useState<string[]>([]);
  const dispatch = useAppDispatch();
  const MAX_SUBREDDITS = planDetails.subscription?.maxSources as number;

  const {
    handleSubmit,
    watch,
    setValue,
    setError,
    register,
    formState: { errors }
  } = useForm<SubredditFormValues>({
    defaultValues: {
      sources: project?.sources ?? [],
      newSubreddit: ""
    },
  });

  const sources = watch("sources");

  const handleAddSubreddit = async (subredditName: string) => {
    const trimmed = subredditName.trim();
    if (!trimmed) return;

    const plainName = trimmed.replace(/^r\//i, "");
    const nameToSend = `r/${plainName}`;

    if (sources.some((s) => s.name.toLowerCase() === plainName.toLowerCase()) || pendingSources.includes(plainName.toLowerCase())) {
      toast({
        title: "Error",
        description: `${nameToSend} is already being tracked`,
        variant: "destructive"
      });
      return;
    }

    setPendingSources((prev) => [...prev, plainName.toLowerCase()]);
    setLoadingSubredditId(nameToSend);

    try {
      const result = await portalClient.addSource({ name: nameToSend });
      const updatedSources = [...sources, result];

      setValue("sources", updatedSources);
      if (project) {
        dispatch(setProject({ ...project, sources: updatedSources }));
      }
      setValue("newSubreddit", "");
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Failed to add";
      toast({
        title: "Error",
        description: message,
        variant: "destructive"
      });
    } finally {
      setPendingSources((prev) =>
        prev.filter((name) => name !== plainName.toLowerCase())
      );
      setLoadingSubredditId(null);
    }
  };

  const handleRemoveSubreddit = async (source: Source) => {
    setLoadingSubredditId(source.id);
    try {
      await portalClient.removeSource({ id: source.id });

      const updatedSources = sources.filter((item) => item.id !== source.id);

      setValue("sources", updatedSources);
      // toast.success(`Removed r/${source.name}`);
      if (project) {
        dispatch(setProject({ ...project, sources: updatedSources }));
      }

    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Failed to remove";
      toast({
        title: "Error",
        description: message,
        variant: "destructive"
      });
    } finally {
      setLoadingSubredditId(null);
    }
  };

  const handleFinish = () => {
    if (sources.length === 0) {
      setError("newSubreddit", { message: "Please add at least one subreddit" });
      return;
    }

    if (!project) return;
    setIsLoading(true);

    try {
      dispatch(setProject(project));
      dispatch(setIsOnboardingDone(true));
      dispatch(finishOnboarding());
      setUser(prev => prev ? {
        ...prev,
        isOnboardingDone: true,
        projects: [...prev.projects, project]
      } : null);
      routers.push(routes.new.dashboard);
    } catch (err: any) {
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

  const availableSuggestions = listOfSuggestedSources.filter((subreddit) => {
    const plainName = subreddit.replace(/^r\//i, "").toLowerCase();
    return !sources.some((s) => s.name.toLowerCase() === plainName);
  });

  const formatSubreddit = (input: string) => {
    const formatted = input.trim().toLowerCase();
    if (formatted.startsWith("r/")) {
      return formatted;
    }
    if (formatted.startsWith("/r/")) {
      return formatted.substring(1);
    }
    if (!formatted.startsWith("r/")) {
      return `r/${formatted}`;
    }
    return formatted;
  };

  const onSubmit = async (data: SubredditFormValues) => {
    await handleAddSubreddit(data.newSubreddit);
  };

  return (
    <div className="space-y-6">
      {/* Current Subreddits */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label>Your Subreddits ({sources.length}/{MAX_SUBREDDITS})</Label>
          <Badge variant="secondary">{MAX_SUBREDDITS - sources.length} remaining</Badge>
        </div>

        {sources.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {sources.map((subreddit) => (
              <Badge key={subreddit.id} variant="default" className={`flex items-center gap-1 ${loadingSubredditId === subreddit.id ? "opacity-25" : ""}`}>
                {subreddit.name}
                <button
                  onClick={() => handleRemoveSubreddit(subreddit)}
                  className="ml-1 hover:bg-primary-foreground/20 rounded-full p-0.5"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : (
          <p className="text-muted-foreground text-sm">No subreddits added yet</p>
        )}
      </div>

      {/* Add New Subreddit */}
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-2">
          <Label htmlFor="newSubreddit">Add Subreddit</Label>
          <div className="flex gap-2">
            <Input
              id="newSubreddit"
              {...register("newSubreddit", {
                validate: (value) => {
                  const formatted = formatSubreddit(value);

                  if (!formatted || formatted === "r/") {
                    return "Please enter a subreddit name";
                  }

                  if (sources.map(item => item.name).includes(formatted)) {
                    return "This subreddit is already added";
                  }

                  if (sources.length >= MAX_SUBREDDITS) {
                    return `You can only add up to ${MAX_SUBREDDITS} subreddits`;
                  }

                  return true;
                },
              })}
              placeholder="Enter subreddit (e.g., entrepreneur or r/entrepreneur)"
              className={errors.newSubreddit?.message ? "border-destructive" : ""}
              disabled={sources.length >= MAX_SUBREDDITS || loadingSubredditId !== null}
            />
            <Button
              type="submit"
              disabled={sources.length >= MAX_SUBREDDITS || loadingSubredditId !== null}
              size="icon"
            >
              <Plus className="w-4 h-4" />
            </Button>
          </div>
          {errors.newSubreddit?.message && <p className="text-sm text-destructive">{errors.newSubreddit?.message}</p>}
          <p className="text-sm text-muted-foreground">
            You can enter with or without the "r/" prefix
          </p>
        </div>
      </form>

      {/* Suggested Subreddits */}
      {availableSuggestions.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Lightbulb className="w-4 h-4" />
              Popular Subreddits for Business
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {availableSuggestions.map((suggestion) => (
                <Button
                  key={suggestion}
                  variant="outline"
                  size="sm"
                  onClick={() => handleAddSubreddit(suggestion)}
                  disabled={sources.length >= MAX_SUBREDDITS || loadingSubredditId === suggestion}
                  className="h-8"
                >
                  <Plus className="w-3 h-3 mr-1" />
                  {suggestion}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Final Step Info */}
      <Card className="bg-primary/5 border-primary/20">
        <CardContent className="pt-6">
          <div className="text-center space-y-2">
            <h3 className="font-semibold">Almost Done! 🎉</h3>
            <p className="text-sm text-muted-foreground">
              Once you complete this step, you'll be ready to start finding leads on Reddit.
              You can always edit these settings later from your dashboard.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex justify-between">
        <Button variant="outline" onClick={() => dispatch(prevStep())}>
          Back to Keywords
        </Button>
        <Button onClick={handleFinish} disabled={isLoading || sources.length === 0} className="bg-green-600 hover:bg-green-700">
          Complete Setup
        </Button>
      </div>
    </div>
  );
}
