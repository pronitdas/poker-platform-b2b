import React from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Typography, Button, Box } from '@mui/material';

const clubs = [
  { id: '1', name: 'Royal Flush Club', status: 'active', players: 234, tables: 12, revenue: '$45,678' },
  { id: '2', name: 'Poker Kings', status: 'active', players: 189, tables: 8, revenue: '$32,456' },
  { id: '3', name: 'Card Sharks', status: 'suspended', players: 156, tables: 6, revenue: '$28,765' },
];

export default function Clubs() {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Clubs</Typography>
        <Button variant="contained">Add Club</Button>
      </Box>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Players</TableCell>
              <TableCell>Tables</TableCell>
              <TableCell>Revenue</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {clubs.map((club) => (
              <TableRow key={club.id}>
                <TableCell>{club.name}</TableCell>
                <TableCell>{club.status}</TableCell>
                <TableCell>{club.players}</TableCell>
                <TableCell>{club.tables}</TableCell>
                <TableCell>{club.revenue}</TableCell>
                <TableCell>
                  <Button size="small">Edit</Button>
                  <Button size="small" color="error">Suspend</Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
