import { useAppDispatch } from "@/store/hooks";
import { setLeadList, setPageNo } from "@/store/Lead/leadSlice";
import { defaultPageNumber } from "@/utils/constants";

export const useSetLeadFilters = () => {
  const dispatch = useAppDispatch();

  
  const resetData = () => {
    dispatch(setLeadList([]));
    dispatch(setPageNo(defaultPageNumber));
  };

  return { resetData };
};
