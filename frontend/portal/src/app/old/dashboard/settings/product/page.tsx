'use client'

import { Box } from '@mui/system'
import { AppBar, Toolbar } from '@mui/material'
import OnboadingForm from '../../../../../components/Onboarding/MainForm';

export default function Page() {

    return (<>
        <Box component='main' sx={{ flexGrow: 1, p: 0, display: 'flex', flexDirection: 'column' }}>
            <AppBar position='static' color='inherit' elevation={0} sx={{ borderBottom: '1px solid #e0e0e0', height: 61 }}>
                <Toolbar>
                    <Box sx={{ flexGrow: 1 }} />
                </Toolbar>
            </AppBar>

            <Box sx={{ p: 3, flexGrow: 1 }}>
                <OnboadingForm />
            </Box>
        </Box>
    </>);
}
