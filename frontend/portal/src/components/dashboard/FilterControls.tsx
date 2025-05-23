
// import { Filter } from "lucide-react";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { DateRangeFilter } from "@doota/pb/doota/portal/v1/portal_pb";

type TimeRangeSelectProps = {
  dateRange?: DateRangeFilter;
  onDateRangeFilterChange?: (value: string) => void;
  leadStatusFilter?: LeadStatus | null;
  onLeadStatusFilterChange?: (value: string) => void;
};

export function FilterControls({ dateRange, leadStatusFilter, onDateRangeFilterChange, onLeadStatusFilterChange }: TimeRangeSelectProps) {

  return (
    <div className="flex flex-wrap gap-2 items-center">
      <div className="relative">
        <select
          className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1"
          value={leadStatusFilter ?? ""}
          onChange={(event) => {
            if (onLeadStatusFilterChange) {
              onLeadStatusFilterChange(event.target.value)
            }
          }}
        >
          <option value="0">All Posts</option>
          <option value="1">Saved Only</option>
        </select>
      </div>

      <div className="relative">
        <select
          className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1"
          value={dateRange}
          onChange={(event) => {
            if (onDateRangeFilterChange) {
              onDateRangeFilterChange(event.target.value)
            }
          }}
        >
          <option value="1">Today</option>
          <option value="2">Yesterday</option>
          <option value="3">Last 7 days</option>
        </select>
      </div>

      {/* <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="all">All Scores</option>
          <option value="high">0.9+ Only</option>
          <option value="medium">0.7-0.9</option>
          <option value="low">Below 0.7</option>
        </select>
      </div> */}

      {/* <div className="relative">
        <select className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1">
          <option value="all">All Tags</option>
          <option value="recommendation">Recommendation</option>
          <option value="pain">Pain Point</option>
          <option value="tools">Looking for Tools</option>
        </select>
      </div> */}

      {/* <button className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium h-9 px-3 border border-input bg-background hover:bg-accent hover:text-accent-foreground">
        <Filter className="h-4 w-4 mr-2" />
        More Filters
      </button> */}
    </div>
  );
}
