"use client"

import ListRenderComp from "./LeadListComp";
import { useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";

const DiscardedTabComponent = () => {
    const { discardedTabList, isLoading } = useAppSelector((state: RootState) => state.lead);

    return (<ListRenderComp list={discardedTabList} isLoading={isLoading} />);
};

export default DiscardedTabComponent;
