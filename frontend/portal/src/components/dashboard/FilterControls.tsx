// import { Filter } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setDateRange, setLeadStatusFilter } from "@/store/Lead/leadSlice";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { DateRangeFilter } from "@doota/pb/doota/portal/v1/portal_pb";

const dateRangeOptions: Record<string, { label: string; value: DateRangeFilter }> = {
  "1": { label: "Today", value: DateRangeFilter.DATE_RANGE_TODAY },
  "2": { label: "Yesterday", value: DateRangeFilter.DATE_RANGE_YESTERDAY },
  "3": { label: "Last 7 days", value: DateRangeFilter.DATE_RANGE_7_DAYS },
};

const leadStatusOptions: Record<string, { label: string; value: LeadStatus }> = {
  "0": { label: "New", value: LeadStatus.NEW },
  "1": { label: "Responded", value: LeadStatus.COMPLETED },
  "2": { label: "Skipped", value: LeadStatus.NOT_RELEVANT },
  "3": { label: "Saved", value: LeadStatus.LEAD },
};

function getKeyFromEnum<T>(options: Record<string, { label: string; value: T }>, value: T): string | undefined {
  return Object.entries(options).find(([, opt]) => opt.value === value)?.[0];
}

function DropdownFilter<T>({ options, selectedValue, onChange }: { options: Record<string, { label: string; value: T }>; selectedValue: T; onChange: (value: T) => void; }) {
  const selectedKey = getKeyFromEnum(options, selectedValue);

  return (
    <div className="relative">
      <select
        className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-1"
        value={selectedKey}
        onChange={(e) => {
          const selected = options[e.target.value];
          if (selected) onChange(selected.value);
        }}
      >
        {Object.entries(options).map(([key, { label }]) => (
          <option key={key} value={key}>
            {label}
          </option>
        ))}
      </select>
    </div>
  );
}

export function FilterControls({ isLeadStatusFilter = true }: { isLeadStatusFilter?: boolean }) {

  const { dateRange, leadStatusFilter } = useAppSelector((state) => state.lead);
  const dispatch = useAppDispatch();

  return (
    <div className="flex flex-wrap gap-2 items-center">
      {isLeadStatusFilter && (
        <DropdownFilter
          options={leadStatusOptions}
          selectedValue={leadStatusFilter}
          onChange={(val) => dispatch(setLeadStatusFilter(val))}
        />
      )}

      <DropdownFilter
        options={dateRangeOptions}
        selectedValue={dateRange}
        onChange={(val) => dispatch(setDateRange(val))}
      />

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
