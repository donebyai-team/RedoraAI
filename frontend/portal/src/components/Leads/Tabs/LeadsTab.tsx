"use client"

import ListRenderComp from "./LeadListComp";
import { useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";

const LeadsTabComponent = () => {
    const { leadsTabList, isLoading } = useAppSelector((state: RootState) => state.lead);

    return (<ListRenderComp list={leadsTabList} isLoading={isLoading} />);
};

export default LeadsTabComponent;
