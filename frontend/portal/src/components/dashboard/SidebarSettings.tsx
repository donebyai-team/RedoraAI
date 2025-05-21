/* eslint-disable jsx-a11y/no-static-element-interactions */
/* eslint-disable jsx-a11y/click-events-have-key-events */

import { Card, CardContent } from "@/components/ui/card";
import {
  // Edit, 
  X,
  Sliders,
  Gauge
} from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { RootState } from "@/store/store";
import { SourceTyeps } from "@/store/Source/sourceSlice";
import { setSubReddit } from "@/store/Params/ParamsSlice";

interface SidebarSettingsProps {
  type: 'keywords' | 'subreddits';
}

export function SidebarSettings({ type }: SidebarSettingsProps) {

  const dispatch = useAppDispatch();
  const project = useAppSelector((state: RootState) => state.stepper.project);
  const { subReddit } = useAppSelector((state: RootState) => state.parems);

  const handleSubRedditsClick = (data: SourceTyeps): void => {
    const subRedditId = data.id === subReddit ? "" : data.id;
    dispatch(setSubReddit(subRedditId));
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Tracked {type === 'keywords' ? 'Keywords' : 'Subreddits'}</h3>
        {/* <button className="text-sm text-primary hover:underline">+ Add New</button> */}
      </div>

      <div className="space-y-2">
        {type === 'keywords' ? (
          // Keywords list
          project?.keywords?.map(item => (
            <div key={item.id} className="flex justify-between items-center p-2 rounded-md hover:bg-secondary/50">
              <div>
                <p className="text-sm font-medium">"{item.name}"</p>
                <p className="text-xs text-muted-foreground">30 matches this week</p>
              </div>
              <div className="flex gap-1">
                {/* <button className="p-1 rounded-md hover:bg-background">
                  <Edit className="h-4 w-4 text-muted-foreground" />
                </button> */}
                <button className="p-1 rounded-md hover:bg-background">
                  <X className="h-4 w-4 text-muted-foreground" />
                </button>
              </div>
            </div>
          ))
        ) : (
          // Subreddits list
          project?.sources?.map(item => (
            <div key={item.id} className={`flex justify-between items-center p-2 rounded-md hover:bg-secondary/50 ${subReddit === item.id ? "bg-secondary/50" : ""}`} onClick={() => handleSubRedditsClick(item)}>
              <div>
                <p className="text-sm font-medium">{item.name}</p>
                <p className="text-xs text-muted-foreground">Last lead: {"3h ago"}</p>
              </div>
              <div className="flex gap-1">
                {/* <button className="p-1 rounded-md hover:bg-background">
                  <Edit className="h-4 w-4 text-muted-foreground" />
                </button> */}
                <button className="p-1 rounded-md hover:bg-background">
                  <X className="h-4 w-4 text-muted-foreground" />
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export function RelevancyScoreSidebar() {
  // Sample current relevancy score
  const currentRelevancyScore = 0.75;

  // Function to determine color based on score
  const getScoreColor = (score: number) => {
    if (score >= 0.9) return "text-green-500";
    if (score >= 0.7) return "text-amber-500";
    return "text-red-500";
  };

  // Function to get text description based on score
  const getScoreDescription = (score: number) => {
    if (score >= 0.9) return "Excellent match";
    if (score >= 0.7) return "Good match";
    if (score >= 0.5) return "Moderate match";
    return "Poor match";
  };

  return (
    <Card className="mt-6">
      <CardContent className="p-6">
        <div className="flex items-center gap-2 mb-4">
          <Gauge className="h-5 w-5 text-primary" />
          <h3 className="text-lg font-medium">Relevancy Score</h3>
        </div>

        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <span className="text-sm font-medium">Current threshold:</span>
            <span className={`text-lg font-bold ${getScoreColor(currentRelevancyScore)}`}>
              {currentRelevancyScore.toFixed(1)}
            </span>
          </div>

          <div className="bg-secondary/30 rounded-lg p-3">
            <p className="text-sm">
              <span className={`font-semibold ${getScoreColor(currentRelevancyScore)}`}>
                {getScoreDescription(currentRelevancyScore)}
              </span>
              : Posts with this score are highly relevant to your target audience.
            </p>
          </div>

          <div className="pt-2">
            <div className="flex justify-between text-xs text-muted-foreground mb-1">
              <span>0.0</span>
              <span>0.5</span>
              <span>1.0</span>
            </div>
            <div className="h-2 w-full bg-secondary rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-red-500 via-amber-500 to-green-500"
                style={{ width: '100%' }}
              />
            </div>
          </div>

          <div className="flex items-center justify-center">
            <button className="flex items-center gap-1 text-sm text-primary hover:underline">
              <Sliders className="h-4 w-4" />
              <span>Adjust settings</span>
            </button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
