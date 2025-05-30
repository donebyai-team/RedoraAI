'use client'

import { Card, Paper, CardContent, Typography } from "@mui/material";
import { Box } from '@mui/system'

export default function Page() {

    return (<>
        <Box component="main" sx={{ flexGrow: 1, p: 0, display: "flex", flexDirection: "column" }}>

            <Box sx={{ p: 3, flexGrow: 1 }}>
                <Card sx={{ p: 4, mt: 5 }} component={Paper}>
                    <CardContent>

                        <Typography variant="h6" color="text.secondary" sx={{ mb: 4 }}>
                            Connect via chat to UPGRADE
                        </Typography>
                    </CardContent>
                </Card>
            </Box>
        </Box>
    </>);
}
