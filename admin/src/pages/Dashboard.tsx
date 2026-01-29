import React from 'react';
import { Grid, Card, CardContent, Typography, Chip, Box } from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

const revenueData = [
  { date: 'Mon', revenue: 4000 },
  { date: 'Tue', revenue: 3000 },
  { date: 'Wed', revenue: 2000 },
  { date: 'Thu', revenue: 2780 },
  { date: 'Fri', revenue: 1890 },
  { date: 'Sat', revenue: 2390 },
  { date: 'Sun', revenue: 3490 },
];

const stats = [
  { title: 'Active Players', value: '1,234', change: '+12%' },
  { title: 'Active Tables', value: '89', change: '+5%' },
  { title: "Today's Revenue", value: '$12,345', change: '+8%' },
  { title: 'Hands Played', value: '45,678', change: '+15%' },
];

export default function Dashboard() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>Dashboard</Typography>
      <Grid container spacing={3}>
        {stats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>{stat.title}</Typography>
                <Typography variant="h4">{stat.value}</Typography>
                <Chip label={stat.change} color="success" size="small" sx={{ mt: 1 }} />
              </CardContent>
            </Card>
          </Grid>
        ))}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Revenue Trend</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={revenueData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <Tooltip />
                  <Line type="monotone" dataKey="revenue" stroke="#1976d2" />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Quick Actions</Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
                <Typography variant="body2">• Create New Club</Typography>
                <Typography variant="body2">• Manage Tables</Typography>
                <Typography variant="body2">• View Reports</Typography>
                <Typography variant="body2">• Player Management</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
