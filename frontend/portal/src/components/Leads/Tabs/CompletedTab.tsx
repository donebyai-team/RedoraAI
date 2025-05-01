"use client"

import ListRenderComp from "./LeadListComp";
import { useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";

const CompletedTabComponent = () => {
    const { completedTabList, isLoading } = useAppSelector((state: RootState) => state.lead);

    return (<ListRenderComp list={completedTabList} isLoading={isLoading} />);
};

export default CompletedTabComponent;
