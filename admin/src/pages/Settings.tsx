import React from 'react';
import { Box, Typography, TextField, Button, Switch, FormControlLabel } from '@mui/material';

export default function Settings() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>Settings</Typography>
      <Box sx={{ maxWidth: 600 }}>
        <Typography variant="h6" gutterBottom>General</Typography>
        <TextField fullWidth label="Platform Name" sx={{ mb: 2 }} />
        <TextField fullWidth label="Support Email" sx={{ mb: 2 }} />
        <Typography variant="h6" gutterBottom>Security</Typography>
        <FormControlLabel control={<Switch />} label="Two-Factor Authentication" sx={{ mb: 2 }} />
        <FormControlLabel control={<Switch />} label="Audit Logging" sx={{ mb: 2 }} />
        <Button variant="contained">Save Settings</Button>
      </Box>
    </Box>
  );
}
