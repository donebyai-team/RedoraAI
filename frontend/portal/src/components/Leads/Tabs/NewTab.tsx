"use client"

import { formatDistanceToNow } from 'date-fns';
import { Timestamp } from "@bufbuild/protobuf/wkt";
import ListRenderComp from "./LeadListComp";
import { SourceTyeps } from "../../../../store/Source/sourceSlice";
import { useAppSelector } from '../../../../store/hooks';
import { RootState } from '../../../../store/store';

export const formateDate = (timestamp: Timestamp): string => {
    const millis = Number(timestamp.seconds) * 1000; // convert bigint to number
    const date = new Date(millis);
    const timeAgo = formatDistanceToNow(date, { addSuffix: true });
    return timeAgo;
};

export const isSameDay = (timestamp: Timestamp): boolean => {
    const date = new Date(Number(timestamp.seconds) * 1000);
    const now = new Date();

    return (
        date.getDate() === now.getDate() &&
        date.getMonth() === now.getMonth() &&
        date.getFullYear() === now.getFullYear()
    );
};

export const getSubredditName = (list: SourceTyeps[], id: string) => {
    const name = list?.find(reddit => reddit.id === id)?.name ?? "N/A";
    return name;
};

export const setLeadActive = (parems_id: string, id: string) => {
    if (parems_id === id) {
        return ({ border: "1px solid #000", backgroundColor: "#0f172a0d", borderRadius: "0.5rem" });
    }
    return ({});
};

const NewTabComponent = () => {

    const { newTabList, isLoading } = useAppSelector((state: RootState) => state.lead);

    return (<ListRenderComp list={newTabList} isLoading={isLoading} />);
};

export default NewTabComponent;
