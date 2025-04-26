"use client";

import { Box } from '@mui/material'
import InboxComponent from '../../../../components/Leads/Inbox'
import LeadsPostDetails from '../../../../components/Leads/LeadsDetails'
import { useState } from 'react';
import { RedditLead } from '@doota/pb/doota/reddit/v1/reddit_pb';

function CombineComp() {
    const [selectedleadData, setSelectedLeadData] = useState<RedditLead | null>(null);

    return (
        <Box className={'flex flex-1 h-full'}>
            <InboxComponent selectedleadData={selectedleadData} setSelectedLeadData={setSelectedLeadData} />
            <LeadsPostDetails selectedleadData={selectedleadData} setSelectedLeadData={setSelectedLeadData} />
        </Box>
    )
}
export default CombineComp;