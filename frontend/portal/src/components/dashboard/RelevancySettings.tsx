
import React, { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Form, FormField, FormItem, FormLabel, FormControl, FormDescription } from "@/components/ui/form";
// import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useForm } from "react-hook-form";
import { toast } from "@/components/ui/use-toast";

interface RelevancyFormValues {
  relevancyThreshold: number;
  autoComment: boolean;
  autoDm: boolean;
  autoSave: boolean;
  commentThreshold: number;
  dmThreshold: number;
  saveThreshold: number;
}

export function RelevancySettings() {
  const [isExpanded, setIsExpanded] = useState(false);
  
  const form = useForm<RelevancyFormValues>({
    defaultValues: {
      relevancyThreshold: 0.7,
      autoComment: false,
      autoDm: false,
      autoSave: true,
      commentThreshold: 0.9,
      dmThreshold: 0.85,
      saveThreshold: 0.75,
    },
  });

  const relevancyThreshold = form.watch("relevancyThreshold");
  const autoComment = form.watch("autoComment");
  const autoDm = form.watch("autoDm");
  const autoSave = form.watch("autoSave");

  const onSubmit = (data: RelevancyFormValues) => {
    // console.log("Saving relevancy settings:", data);
    toast({
      title: "Settings saved",
      description: "Your relevancy and automation settings have been updated.",
    });
  };

  const getThresholdColor = (value: number): string => {
    if (value >= 0.9) return "text-green-500";
    if (value >= 0.7) return "text-amber-500";
    return "text-red-500";
  };

  return (
    <Card className="mb-6">
      <CardContent className="p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Relevancy & Automation Settings</h3>
          <Button 
            variant="outline" 
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? "Hide Details" : "Show Details"}
          </Button>
        </div>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <FormLabel>Minimum Relevancy Score</FormLabel>
                <span className={`font-bold ${getThresholdColor(relevancyThreshold)}`}>
                  {relevancyThreshold.toFixed(2)}
                </span>
              </div>
              <Slider
                value={[relevancyThreshold * 100]} 
                min={0}
                max={100}
                step={5}
                onValueChange={(value) => form.setValue("relevancyThreshold", value[0] / 100)}
                className="w-full"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Posts below this score won't appear in your dashboard
              </p>
            </div>

            {isExpanded && (
              <div className="space-y-6 border-t pt-4 mt-4">
                <h4 className="font-medium">Automation Settings</h4>
                
                <FormField
                  control={form.control}
                  name="autoSave"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                      <div className="space-y-0.5">
                        <FormLabel>Auto-Save Relevant Posts</FormLabel>
                        <FormDescription>
                          Automatically save posts above threshold
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />
                
                {autoSave && (
                  <div className="ml-6 space-y-2">
                    <FormLabel className="text-sm">Save Threshold</FormLabel>
                    <div className="flex items-center gap-2">
                      <Slider
                        value={[form.watch("saveThreshold") * 100]} 
                        min={0}
                        max={100}
                        step={5}
                        onValueChange={(value) => form.setValue("saveThreshold", value[0] / 100)}
                        className="flex-grow"
                      />
                      <span className={`w-12 font-medium ${getThresholdColor(form.watch("saveThreshold"))}`}>
                        {form.watch("saveThreshold").toFixed(2)}
                      </span>
                    </div>
                  </div>
                )}

                <FormField
                  control={form.control}
                  name="autoComment"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                      <div className="space-y-0.5">
                        <FormLabel>Auto-Comment</FormLabel>
                        <FormDescription>
                          Automatically post AI-suggested comments
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />
                
                {autoComment && (
                  <div className="ml-6 space-y-2">
                    <FormLabel className="text-sm">Comment Threshold</FormLabel>
                    <div className="flex items-center gap-2">
                      <Slider
                        value={[form.watch("commentThreshold") * 100]} 
                        min={0}
                        max={100}
                        step={5}
                        onValueChange={(value) => form.setValue("commentThreshold", value[0] / 100)}
                        className="flex-grow"
                      />
                      <span className={`w-12 font-medium ${getThresholdColor(form.watch("commentThreshold"))}`}>
                        {form.watch("commentThreshold").toFixed(2)}
                      </span>
                    </div>
                  </div>
                )}
                
                <FormField
                  control={form.control}
                  name="autoDm"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                      <div className="space-y-0.5">
                        <FormLabel>Auto-DM</FormLabel>
                        <FormDescription>
                          Automatically send AI-suggested DMs
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />
                
                {autoDm && (
                  <div className="ml-6 space-y-2">
                    <FormLabel className="text-sm">DM Threshold</FormLabel>
                    <div className="flex items-center gap-2">
                      <Slider
                        value={[form.watch("dmThreshold") * 100]} 
                        min={0}
                        max={100}
                        step={5}
                        onValueChange={(value) => form.setValue("dmThreshold", value[0] / 100)}
                        className="flex-grow"
                      />
                      <span className={`w-12 font-medium ${getThresholdColor(form.watch("dmThreshold"))}`}>
                        {form.watch("dmThreshold").toFixed(2)}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            )}
            
            <div className="flex justify-end pt-2">
              <Button type="submit">Save Settings</Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
