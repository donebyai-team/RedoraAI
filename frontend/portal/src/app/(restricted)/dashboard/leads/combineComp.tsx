"use client";

import { Box } from '@mui/material'
import InboxComponent from '../../../../components/Leads/Inbox'
import LeadsPostDetails from '../../../../components/Leads/LeadsDetails'

function CombineComp() {
    return (
        <Box className={'flex flex-1 h-full'}>
            <InboxComponent />
            <LeadsPostDetails />
        </Box>
    );
}
export default CombineComp;