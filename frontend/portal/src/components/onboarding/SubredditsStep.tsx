
import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X, Plus, Lightbulb } from "lucide-react";

interface SubredditsStepProps {
  data: string[];
  onUpdate: (subreddits: string[]) => void;
  onFinish: () => void;
  onPrev: () => void;
}

const SUGGESTED_SUBREDDITS = [
  "r/entrepreneur", "r/startups", "r/marketing", "r/smallbusiness",
  "r/SaaS", "r/digitalnomad", "r/growthstrategy", "r/sales",
  "r/productivity", "r/business", "r/freelance", "r/socialmedia",
  "r/ecommerce", "r/webdev", "r/SEO"
];

const MAX_SUBREDDITS = 8; // This could be dynamic based on subscription plan

export default function SubredditsStep({ data, onUpdate, onFinish, onPrev }: SubredditsStepProps) {
  const [subreddits, setSubreddits] = useState<string[]>(data);
  const [newSubreddit, setNewSubreddit] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    onUpdate(subreddits);
  }, [subreddits, onUpdate]);

  const formatSubreddit = (input: string) => {
    let formatted = input.trim().toLowerCase();
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

  const addSubreddit = (subreddit: string) => {
    const formattedSubreddit = formatSubreddit(subreddit);
    
    if (!formattedSubreddit || formattedSubreddit === "r/") {
      setError("Please enter a subreddit name");
      return;
    }

    if (subreddits.includes(formattedSubreddit)) {
      setError("This subreddit is already added");
      return;
    }

    if (subreddits.length >= MAX_SUBREDDITS) {
      setError(`You can only add up to ${MAX_SUBREDDITS} subreddits`);
      return;
    }

    setSubreddits(prev => [...prev, formattedSubreddit]);
    setNewSubreddit("");
    setError("");
  };

  const removeSubreddit = (subredditToRemove: string) => {
    setSubreddits(prev => prev.filter(subreddit => subreddit !== subredditToRemove));
    setError("");
  };

  const addSuggestedSubreddit = (subreddit: string) => {
    addSubreddit(subreddit);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addSubreddit(newSubreddit);
    }
  };

  const handleFinish = () => {
    if (subreddits.length === 0) {
      setError("Please add at least one subreddit");
      return;
    }
    onFinish();
  };

  const availableSuggestions = SUGGESTED_SUBREDDITS.filter(
    suggestion => !subreddits.includes(suggestion.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Current Subreddits */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label>Your Subreddits ({subreddits.length}/{MAX_SUBREDDITS})</Label>
          <Badge variant="secondary">{MAX_SUBREDDITS - subreddits.length} remaining</Badge>
        </div>
        
        {subreddits.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {subreddits.map((subreddit) => (
              <Badge key={subreddit} variant="default" className="flex items-center gap-1">
                {subreddit}
                <button
                  onClick={() => removeSubreddit(subreddit)}
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
      <div className="space-y-2">
        <Label htmlFor="newSubreddit">Add Subreddit</Label>
        <div className="flex gap-2">
          <Input
            id="newSubreddit"
            value={newSubreddit}
            onChange={(e) => {
              setNewSubreddit(e.target.value);
              setError("");
            }}
            onKeyPress={handleKeyPress}
            placeholder="Enter subreddit (e.g., entrepreneur or r/entrepreneur)"
            className={error ? "border-destructive" : ""}
            disabled={subreddits.length >= MAX_SUBREDDITS}
          />
          <Button
            type="button"
            onClick={() => addSubreddit(newSubreddit)}
            disabled={subreddits.length >= MAX_SUBREDDITS}
            size="icon"
          >
            <Plus className="w-4 h-4" />
          </Button>
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <p className="text-sm text-muted-foreground">
          You can enter with or without the "r/" prefix
        </p>
      </div>

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
              {availableSuggestions.slice(0, 8).map((suggestion) => (
                <Button
                  key={suggestion}
                  variant="outline"
                  size="sm"
                  onClick={() => addSuggestedSubreddit(suggestion)}
                  disabled={subreddits.length >= MAX_SUBREDDITS}
                  className="h-8"
                >
                  <Plus className="w-3 h-3 mr-1" />
                  {suggestion}
                </Button>
              ))}
            </div>
            {availableSuggestions.length > 8 && (
              <p className="text-sm text-muted-foreground mt-2">
                And {availableSuggestions.length - 8} more suggestions...
              </p>
            )}
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
        <Button variant="outline" onClick={onPrev}>
          Back to Keywords
        </Button>
        <Button onClick={handleFinish} className="bg-green-600 hover:bg-green-700">
          Complete Setup
        </Button>
      </div>
    </div>
  );
}
