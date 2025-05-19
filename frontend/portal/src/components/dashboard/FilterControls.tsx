
import { Filter } from "lucide-react";

export function FilterControls() {
  return (
    <div className="flex flex-wrap gap-2 items-center">
      <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="all">All Posts</option>
          <option value="unreplied">Unreplied Only</option>
          <option value="saved">Saved Only</option>
        </select>
      </div>
      
      <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="today">Today</option>
          <option value="7days">Last 7 days</option>
          <option value="30days">Last 30 days</option>
        </select>
      </div>
      
      <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="all">All Scores</option>
          <option value="high">0.9+ Only</option>
          <option value="medium">0.7-0.9</option>
          <option value="low">Below 0.7</option>
        </select>
      </div>
      
      <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="all">All Tags</option>
          <option value="recommendation">Recommendation</option>
          <option value="pain">Pain Point</option>
          <option value="tools">Looking for Tools</option>
        </select>
      </div>
      
      <button className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium h-9 px-3 border border-input bg-background hover:bg-accent hover:text-accent-foreground">
        <Filter className="h-4 w-4 mr-2" />
        More Filters
      </button>
    </div>
  );
}
