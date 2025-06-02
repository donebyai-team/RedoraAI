
import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X, Plus, Lightbulb } from "lucide-react";
import { toast } from "@/hooks/use-toast";

interface KeywordsStepProps {
  data: string[];
  onUpdate: (keywords: string[]) => void;
  onNext: () => void;
  onPrev: () => void;
}

const SUGGESTED_KEYWORDS = [
  "lead generation", "marketing automation", "CRM", "sales funnel", 
  "customer acquisition", "email marketing", "conversion optimization",
  "growth hacking", "digital marketing", "business tools", "SaaS",
  "productivity tools", "analytics", "A/B testing", "customer retention"
];

const MAX_KEYWORDS = 10; // This could be dynamic based on subscription plan

export default function KeywordsStep({ data, onUpdate, onNext, onPrev }: KeywordsStepProps) {
  const [keywords, setKeywords] = useState<string[]>(data);
  const [newKeyword, setNewKeyword] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    onUpdate(keywords);
  }, [keywords, onUpdate]);

  const addKeyword = (keyword: string) => {
    const trimmedKeyword = keyword.trim().toLowerCase();
    
    if (!trimmedKeyword) {
      setError("Please enter a keyword");
      return;
    }

    if (keywords.includes(trimmedKeyword)) {
      setError("This keyword is already added");
      return;
    }

    if (keywords.length >= MAX_KEYWORDS) {
      setError(`You can only add up to ${MAX_KEYWORDS} keywords`);
      return;
    }

    setKeywords(prev => [...prev, trimmedKeyword]);
    setNewKeyword("");
    setError("");
  };

  const removeKeyword = (keywordToRemove: string) => {
    setKeywords(prev => prev.filter(keyword => keyword !== keywordToRemove));
    setError("");
  };

  const addSuggestedKeyword = (keyword: string) => {
    addKeyword(keyword);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addKeyword(newKeyword);
    }
  };

  const handleNext = () => {
    if (keywords.length === 0) {
      setError("Please add at least one keyword");
      return;
    }
    onNext();
  };

  const availableSuggestions = SUGGESTED_KEYWORDS.filter(
    suggestion => !keywords.includes(suggestion.toLowerCase())
  );

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
            {keywords.map((keyword) => (
              <Badge key={keyword} variant="default" className="flex items-center gap-1">
                {keyword}
                <button
                  onClick={() => removeKeyword(keyword)}
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
      <div className="space-y-2">
        <Label htmlFor="newKeyword">Add Keyword</Label>
        <div className="flex gap-2">
          <Input
            id="newKeyword"
            value={newKeyword}
            onChange={(e) => {
              setNewKeyword(e.target.value);
              setError("");
            }}
            onKeyPress={handleKeyPress}
            placeholder="Enter a keyword to track"
            className={error ? "border-destructive" : ""}
            disabled={keywords.length >= MAX_KEYWORDS}
          />
          <Button
            type="button"
            onClick={() => addKeyword(newKeyword)}
            disabled={keywords.length >= MAX_KEYWORDS}
            size="icon"
          >
            <Plus className="w-4 h-4" />
          </Button>
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>

      {/* Suggested Keywords */}
      {availableSuggestions.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Lightbulb className="w-4 h-4" />
              Suggested Keywords
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {availableSuggestions.slice(0, 8).map((suggestion) => (
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
            {availableSuggestions.length > 8 && (
              <p className="text-sm text-muted-foreground mt-2">
                And {availableSuggestions.length - 8} more suggestions...
              </p>
            )}
          </CardContent>
        </Card>
      )}

      {/* Navigation */}
      <div className="flex justify-between">
        <Button variant="outline" onClick={onPrev}>
          Back to Product Details
        </Button>
        <Button onClick={handleNext}>
          Continue to Subreddits
        </Button>
      </div>
    </div>
  );
}
