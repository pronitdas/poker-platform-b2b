import React from 'react';
import { Box, Typography, Grid, Card, CardContent, Button } from '@mui/material';

export default function Tables() {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Tables</Typography>
        <Button variant="contained">Create Table</Button>
      </Box>
      <Grid container spacing={3}>
        {[1, 2, 3, 4, 5, 6].map((id) => (
          <Grid item xs={12} sm={6} md={4} key={id}>
            <Card>
              <CardContent>
                <Typography variant="h6">Table {id}</Typography>
                <Typography color="textSecondary">No Limit Texas Hold'em</Typography>
                <Typography variant="body2" sx={{ mt: 2 }}>
                  Blinds: $5/$10
                </Typography>
                <Typography variant="body2">
                  Players: 3/9
                </Typography>
                <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                  <Button size="small" variant="outlined">Edit</Button>
                  <Button size="small" color="error">Close</Button>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
