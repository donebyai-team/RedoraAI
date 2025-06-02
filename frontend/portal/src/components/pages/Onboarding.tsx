

"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Check } from "lucide-react";
import { toast } from "@/hooks/use-toast";
import ProductDetailsStep from "@/components/onboarding/ProductDetailsStep";
import KeywordsStep from "@/components/onboarding/KeywordsStep";
import SubredditsStep from "@/components/onboarding/SubredditsStep";
import { useRouter } from "next/navigation";

export interface OnboardingData {
  productDetails: {
    website: string;
    productName: string;
    description: string;
    targetPersona: string;
  };
  keywords: string[];
  subreddits: string[];
}

const STEPS = [
  {
    id: 1,
    title: "Product Details",
    description: "Tell us about your product and target audience",
    purpose: "We'll use this to find relevant Reddit discussions where your audience might be seeking solutions like yours."
  },
  {
    id: 2,
    title: "Keywords",
    description: "Choose keywords to track in Reddit posts",
    purpose: "These keywords help us identify posts and comments where people are discussing topics related to your product."
  },
  {
    id: 3,
    title: "Subreddits",
    description: "Select communities to monitor for opportunities",
    purpose: "We'll focus our search on these specific Reddit communities where your target audience is most active."
  },
];

export default function Onboarding() {
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState(1);
  const [completedSteps, setCompletedSteps] = useState<number[]>([]);
  const [data, setData] = useState<OnboardingData>({
    productDetails: {
      website: "",
      productName: "",
      description: "",
      targetPersona: "",
    },
    keywords: [],
    subreddits: [],
  });

  const updateData = (stepData: Partial<OnboardingData>) => {
    setData(prev => ({ ...prev, ...stepData }));
  };

  const markStepCompleted = (step: number) => {
    if (!completedSteps.includes(step)) {
      setCompletedSteps(prev => [...prev, step]);
    }
  };

  const goToStep = (step: number) => {
    setCurrentStep(step);
  };

  const nextStep = () => {
    if (currentStep < STEPS.length) {
      markStepCompleted(currentStep);
      setCurrentStep(currentStep + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  const finishOnboarding = () => {
    markStepCompleted(currentStep);
    toast({
      title: "Onboarding completed!",
      description: "Your account is now set up and ready to use.",
    });
    router.push("/dashboard");
  };

  const renderStep = () => {
    switch (currentStep) {
      case 1:
        return (
          <ProductDetailsStep
            data={data.productDetails}
            onUpdate={(productDetails) => updateData({ productDetails })}
            onNext={nextStep}
          />
        );
      case 2:
        return (
          <KeywordsStep
            data={data.keywords}
            onUpdate={(keywords) => updateData({ keywords })}
            onNext={nextStep}
            onPrev={prevStep}
          />
        );
      case 3:
        return (
          <SubredditsStep
            data={data.subreddits}
            onUpdate={(subreddits) => updateData({ subreddits })}
            onFinish={finishOnboarding}
            onPrev={prevStep}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 py-12 px-6">
      <div className="max-w-5xl mx-auto h-full">
        {/* Main Content - Vertical Layout */}
        <div className="flex gap-8 min-h-[calc(100vh-6rem)]">
          {/* Left Side - Vertical Steppers */}
          <div className="w-72 flex-shrink-0">
            <Card className="border-0 shadow-xl bg-white/95 backdrop-blur-sm rounded-2xl h-full">
              <CardContent className="p-6 space-y-6">
                {/* Vertical Steps */}
                {STEPS.map((step, index) => (
                  <div key={step.id} className="relative">
                    <div className="flex items-start gap-3">
                      {/* Step Circle */}
                      <button
                        onClick={() => goToStep(step.id)}
                        disabled={!completedSteps.includes(step.id) && step.id !== currentStep}
                        className={`group relative flex items-center justify-center w-10 h-10 rounded-xl text-sm font-bold transition-all duration-300 flex-shrink-0 ${completedSteps.includes(step.id)
                            ? "bg-gradient-to-r from-green-500 to-emerald-600 text-white shadow-lg shadow-green-200/50"
                            : step.id === currentStep
                              ? "bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-200/50"
                              : "bg-gray-100 text-gray-400 border-2 border-gray-200 hover:border-gray-300"
                          } ${(completedSteps.includes(step.id) || step.id === currentStep)
                            ? "cursor-pointer hover:scale-105"
                            : "cursor-not-allowed"
                          }`}
                      >
                        {completedSteps.includes(step.id) ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <span>{step.id}</span>
                        )}

                        {/* Glow effect for current step */}
                        {step.id === currentStep && !completedSteps.includes(step.id) && (
                          <div className="absolute inset-0 rounded-xl bg-gradient-to-r from-blue-500 to-indigo-600 opacity-30 blur-md animate-pulse"></div>
                        )}
                      </button>

                      {/* Step Content */}
                      <div className="flex-1 min-w-0">
                        <div className={`font-semibold text-sm mb-1 transition-colors ${step.id === currentStep ? "text-blue-600" :
                            completedSteps.includes(step.id) ? "text-green-600" : "text-gray-500"
                          }`}>
                          {step.title}
                        </div>
                        <div className="text-xs text-gray-500 leading-relaxed">
                          {step.description}
                        </div>
                      </div>
                    </div>

                    {/* Vertical connector line */}
                    {index < STEPS.length - 1 && (
                      <div className="flex justify-start ml-5 mt-4 mb-2">
                        <div className={`w-0.5 h-6 rounded-full transition-all duration-500 ${completedSteps.includes(step.id) ?
                            "bg-gradient-to-b from-green-400 to-emerald-500" :
                            step.id === currentStep ?
                              "bg-gradient-to-b from-blue-400 to-indigo-500" :
                              "bg-gray-200"
                          }`}>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>

          {/* Right Side - Step Content */}
          <div className="flex-1">
            <Card className="border-0 shadow-2xl bg-white/95 backdrop-blur-sm rounded-3xl overflow-hidden h-full">
              <div className="bg-gradient-to-r from-blue-500/10 via-indigo-500/10 to-purple-500/10 p-1 h-full">
                <div className="bg-white rounded-3xl h-full flex flex-col">
                  <CardHeader className="pb-4 pt-6 px-8 flex-shrink-0">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-base ${completedSteps.includes(currentStep) ?
                          "bg-gradient-to-r from-green-500 to-emerald-600" :
                          "bg-gradient-to-r from-blue-500 to-indigo-600"
                        }`}>
                        {completedSteps.includes(currentStep) ? (
                          <Check className="w-5 h-5" />
                        ) : (
                          currentStep
                        )}
                      </div>
                      <div>
                        <CardTitle className="text-xl font-bold text-gray-900">
                          {STEPS[currentStep - 1].title}
                        </CardTitle>
                        <CardDescription className="text-sm text-gray-600 mt-1">
                          {STEPS[currentStep - 1].description}
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="px-8 pb-8 flex-1 overflow-auto">
                    <div className="animate-fade-in">
                      {renderStep()}
                    </div>
                  </CardContent>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

