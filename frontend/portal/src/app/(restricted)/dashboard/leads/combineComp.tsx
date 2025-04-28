"use client";

import { Box } from '@mui/material'
import InboxComponent from '../../../../components/Leads/Inbox'
import LeadsPostDetails from '../../../../components/Leads/LeadsDetails'
import { useState } from 'react';
import { Lead } from '@doota/pb/doota/core/v1/core_pb';

function CombineComp() {
    const [selectedleadData, setSelectedLeadData] = useState<Lead | null>(null);

    return (
        <Box className={'flex flex-1 h-full'}>
            <InboxComponent />
            <LeadsPostDetails />
        </Box>
    )
}
export default CombineComp;