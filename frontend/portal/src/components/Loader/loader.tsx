"use client";

import { Box } from "@mui/system";
import { FallbackSpinner } from "@/atoms/FallbackSpinner";

export const AuthLoading = () => (
  <Box
    sx={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      overflowX: 'hidden',
      position: 'relative',
      width: '100%'
    }}
  >
    <FallbackSpinner />
  </Box>
);